import { BadRequestException, Injectable } from "@nestjs/common";
import { JwtService } from "@nestjs/jwt";
import { PrismaService } from "../prisma/prisma.service";
import { OnboardingCompleteDto } from "./dto/onboarding-complete.dto";

@Injectable()
export class OnboardingService {
  constructor(
    private prisma: PrismaService,
    private jwt: JwtService
  ) {}

  async getStatus(userId: string) {
    const user = await this.prisma.user.findUnique({
      where: { id: userId },
      select: {
        onboardingCompleted: true,
        category: true,
        fullName: true,
        grade: true,
        board: true,
        subject: true,
        organizationName: true,
        organizationType: true,
      },
    });
    return user;
  }

  async complete(userId: string, dto: OnboardingCompleteDto) {
    if (dto.grade) {
      const exists = await this.prisma.gradeMaster.findUnique({ where: { id: dto.grade } });
      if (!exists) throw new BadRequestException("Invalid grade");
    }
    if (dto.board) {
      const exists = await this.prisma.boardMaster.findUnique({ where: { id: dto.board } });
      if (!exists) throw new BadRequestException("Invalid board");
    }
    if (dto.subject) {
      const exists = await this.prisma.subjectMaster.findUnique({ where: { id: dto.subject } });
      if (!exists) throw new BadRequestException("Invalid subject");
    }

    const user = await this.prisma.user.update({
      where: { id: userId },
      data: {
        onboardingCompleted: true,
        category: dto.category,
        fullName: dto.fullName,
        grade: dto.grade,
        board: dto.board,
        subject: dto.subject,
        organizationName: dto.organizationName,
        organizationType: dto.organizationType,
        // Also update name if fullName provided
        ...(dto.fullName ? { name: dto.fullName } : {}),
      },
    });

    // Issue new JWT with onboardingCompleted: true
    const payload = { sub: user.id, email: user.email, onboardingCompleted: true, isStarted: user.isStarted };
    return {
      accessToken: this.jwt.sign(payload),
      user: {
        id: user.id,
        email: user.email,
        name: user.name,
        plan: user.plan,
        onboardingCompleted: true,
        isStarted: user.isStarted,
      },
    };
  }
}
