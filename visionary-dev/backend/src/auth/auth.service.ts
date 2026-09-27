import { Injectable, UnauthorizedException, ConflictException, BadRequestException } from "@nestjs/common";
import { JwtService } from "@nestjs/jwt";
import { PrismaService } from "../prisma/prisma.service";
import { EmailService } from "../email/email.service";
import * as bcrypt from "bcrypt";
import { randomInt } from "crypto";

const OTP_LENGTH = 6;
const OTP_TTL_MS = 10 * 60 * 1000;
const OTP_MAX_ATTEMPTS = 5;
const OTP_LOCKOUT_MS = 15 * 60 * 1000;
const OTP_RESEND_COOLDOWN_MS = 60 * 1000;

@Injectable()
export class AuthService {
  constructor(
    private prisma: PrismaService,
    private jwt: JwtService,
    private email: EmailService,
  ) {}

  private generateOtp(): string {
    return randomInt(0, 10 ** OTP_LENGTH).toString().padStart(OTP_LENGTH, "0");
  }

  async register(emailAddr: string, password: string) {
    const existing = await this.prisma.user.findUnique({ where: { email: emailAddr } });
    if (existing) throw new ConflictException("Email already in use");

    const passwordHash = await bcrypt.hash(password, 10);
    const otp = this.generateOtp();
    const otpHash = await bcrypt.hash(otp, 10);
    const otpExpiry = new Date(Date.now() + OTP_TTL_MS);

    await this.prisma.user.create({
      data: {
        email: emailAddr,
        name: "",
        passwordHash,
        otp: otpHash,
        otpExpiry,
        otpAttempts: 0,
        otpLockedUntil: null,
      },
    });

    await this.email.sendOtp(emailAddr, otp);

    return { message: "OTP sent", email: emailAddr };
  }

  async verifyOtp(emailAddr: string, otp: string) {
    const user = await this.prisma.user.findUnique({ where: { email: emailAddr } });
    if (!user) throw new BadRequestException("Invalid OTP");
    if (user.otpLockedUntil && user.otpLockedUntil > new Date()) {
      throw new BadRequestException("Too many attempts. Try again later.");
    }
    if (!user.otp || !user.otpExpiry) throw new BadRequestException("Invalid OTP");
    if (new Date() > user.otpExpiry) throw new BadRequestException("Invalid OTP");

    const matches = await bcrypt.compare(otp, user.otp);
    if (!matches) {
      const incremented = await this.prisma.user.update({
        where: { email: emailAddr },
        data: { otpAttempts: { increment: 1 } },
        select: { otpAttempts: true },
      });
      if (incremented.otpAttempts >= OTP_MAX_ATTEMPTS) {
        await this.prisma.user.update({
          where: { email: emailAddr },
          data: {
            otp: null,
            otpExpiry: null,
            otpAttempts: 0,
            otpLockedUntil: new Date(Date.now() + OTP_LOCKOUT_MS),
          },
        });
        throw new BadRequestException("Too many attempts. Try again later.");
      }
      throw new BadRequestException("Invalid OTP");
    }

    const updated = await this.prisma.user.update({
      where: { email: emailAddr },
      data: {
        isVerified: true,
        otp: null,
        otpExpiry: null,
        otpAttempts: 0,
        otpLockedUntil: null,
      },
    });

    return this.issueToken(updated);
  }

  async resendOtp(emailAddr: string) {
    const user = await this.prisma.user.findUnique({ where: { email: emailAddr } });
    if (!user) throw new BadRequestException("User not found");
    if (user.isVerified) throw new BadRequestException("Already verified");
    if (user.otpLockedUntil && user.otpLockedUntil > new Date()) {
      throw new BadRequestException("Too many attempts. Try again later.");
    }

    if (user.otpExpiry) {
      const previousIssuedAt = user.otpExpiry.getTime() - OTP_TTL_MS;
      const earliestResendAt = previousIssuedAt + OTP_RESEND_COOLDOWN_MS;
      if (Date.now() < earliestResendAt) {
        throw new BadRequestException("Please wait before requesting another OTP.");
      }
    }

    const otp = this.generateOtp();
    const otpHash = await bcrypt.hash(otp, 10);
    const otpExpiry = new Date(Date.now() + OTP_TTL_MS);

    await this.prisma.user.update({
      where: { email: emailAddr },
      data: {
        otp: otpHash,
        otpExpiry,
        otpLockedUntil: null,
      },
    });

    await this.email.sendOtp(emailAddr, otp);

    return { message: "OTP resent" };
  }

  async login(emailAddr: string, password: string) {
    const user = await this.prisma.user.findUnique({ where: { email: emailAddr } });
    if (!user || !user.passwordHash) throw new UnauthorizedException("Invalid credentials");

    const valid = await bcrypt.compare(password, user.passwordHash);
    if (!valid) throw new UnauthorizedException("Invalid credentials");

    if (!user.isVerified) throw new UnauthorizedException("Please verify your email first");

    if (user.isDeleted) throw new UnauthorizedException("ACCOUNT_DELETED");

    return this.issueToken(user);
  }

  async googleLogin(googleUser: { googleId: string; email: string; name: string; avatar?: string }) {
    let user = await this.prisma.user.findUnique({ where: { googleId: googleUser.googleId } });

    if (!user) {
      user = await this.prisma.user.create({
        data: {
          email: googleUser.email,
          name: googleUser.name,
          avatar: googleUser.avatar,
          googleId: googleUser.googleId,
          isVerified: true,
        },
      });
    }

    if (user.isDeleted) throw new UnauthorizedException("ACCOUNT_DELETED");

    return this.issueToken(user);
  }

  async getProfile(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: {
        id: true,
        email: true,
        name: true,
        fullName: true,
        schoolName: true,
        phone: true,
        grade: true,
        board: true,
        onboardingCompleted: true,
        isStarted: true,
      },
    });

    // Resolve grade and board IDs to names for display
    const [gradeRecord, boardRecord] = await Promise.all([
      user?.grade
        ? this.prisma.gradeMaster.findUnique({ where: { id: user.grade }, select: { name: true } })
        : null,
      user?.board
        ? this.prisma.boardMaster.findUnique({ where: { id: user.board }, select: { name: true } })
        : null,
    ]);

    return {
      ...user,
      gradeName: gradeRecord?.name ?? null,
      boardName: boardRecord?.name ?? null,
    };
  }

  async updateProfile(userId: string, dto: { fullName?: string; schoolName?: string; phone?: string }) {
    const data: Record<string, string | null> = {};
    if (dto.fullName  !== undefined) data.fullName  = dto.fullName;
    if (dto.schoolName !== undefined) data.schoolName = dto.schoolName;
    if (dto.phone     !== undefined) data.phone     = dto.phone;

    await this.prisma.user.update({ where: { id: userId }, data });
    return this.getProfile(userId);
  }

  async requestAccountDeletion(userId: string, reason?: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: { email: true, name: true, fullName: true },
    });
    if (!user) throw new BadRequestException("User not found");

    await this.prisma.$transaction([
      this.prisma.deletionRequest.create({
        data: { userId, reason: reason ?? null, status: "pending" },
      }),
      this.prisma.user.update({
        where: { id: userId },
        data: { isDeleted: true },
      }),
    ]);

    await this.email.sendDeletionConfirmation(user.email, user.fullName ?? user.name);

    return { message: "Deletion request submitted" };
  }

  async changePassword(userId: string, currentPassword: string, newPassword: string) {
    const user = await this.prisma.user.findUnique({ where: { id: userId } });
    if (!user || !user.passwordHash) throw new BadRequestException("Password change not available for this account");

    const valid = await bcrypt.compare(currentPassword, user.passwordHash);
    if (!valid) throw new BadRequestException("Current password is incorrect");

    const newHash = await bcrypt.hash(newPassword, 10);
    await this.prisma.user.update({ where: { id: userId }, data: { passwordHash: newHash } });
    return { message: "Password updated successfully" };
  }

  issueToken(user: { id: string; email: string; name: string; plan: string; onboardingCompleted: boolean; isStarted: boolean }) {
    const payload = { sub: user.id, email: user.email, onboardingCompleted: user.onboardingCompleted, isStarted: user.isStarted };
    return {
      accessToken: this.jwt.sign(payload),
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        plan: user.plan,
        onboardingCompleted: user.onboardingCompleted,
        isStarted: user.isStarted,
      },
    };
  }
}
