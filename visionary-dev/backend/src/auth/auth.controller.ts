import { Controller, Post, Patch, Body, Get, UseGuards, Req, Res, HttpCode } from "@nestjs/common";
import type { Response } from "express";
import { AuthService } from "./auth.service";
import { AuthGuard } from "@nestjs/passport";
import { ConfigService } from "@nestjs/config";
import { RegisterDto } from "./dto/register.dto";
import { VerifyOtpDto } from "./dto/verify-otp.dto";
import { ResendOtpDto } from "./dto/resend-otp.dto";
import { LoginDto } from "./dto/login.dto";
import { UpdateProfileDto } from "./dto/update-profile.dto";
import { setAuthCookie, clearAuthCookie } from "./auth.cookie";

@Controller("auth")
export class AuthController {
  constructor(
    private authService: AuthService,
    private config: ConfigService
  ) {}

  @Post("register")
  register(@Body() dto: RegisterDto) {
    return this.authService.register(dto.email, dto.password);
  }

  @Post("verify-otp")
  async verifyOtp(@Body() dto: VerifyOtpDto, @Res({ passthrough: true }) res: Response) {
    const result = await this.authService.verifyOtp(dto.email, dto.otp);
    setAuthCookie(res, result.accessToken);
    return result;
  }

  @Post("resend-otp")
  resendOtp(@Body() dto: ResendOtpDto) {
    return this.authService.resendOtp(dto.email);
  }

  @Post("login")
  async login(@Body() dto: LoginDto, @Res({ passthrough: true }) res: Response) {
    const result = await this.authService.login(dto.email, dto.password);
    setAuthCookie(res, result.accessToken);
    return result;
  }

  @Get("google")
  @UseGuards(AuthGuard("google"))
  googleAuth() {
    // Redirects to Google — handled by passport
  }

  @Get("google/callback")
  @UseGuards(AuthGuard("google"))
  async googleCallback(@Req() req: any, @Res() res: Response) {
    const frontendUrl = this.config.get("FRONTEND_URL") || "http://localhost:3000";
    try {
      const result = await this.authService.googleLogin(req.user);
      setAuthCookie(res, result.accessToken);
      return res.redirect(
        `${frontendUrl}/auth/callback?onboardingCompleted=${result.user.onboardingCompleted}&isStarted=${result.user.isStarted}`
      );
    } catch (err: any) {
      if (err?.message === "ACCOUNT_DELETED") {
        return res.redirect(`${frontendUrl}/login?deleted=1`);
      }
      return res.redirect(`${frontendUrl}/login?error=oauth_failed`);
    }
  }

  @Post("logout")
  @HttpCode(200)
  logout(@Res({ passthrough: true }) res: Response) {
    clearAuthCookie(res);
    return { ok: true };
  }

  @Get("me")
  @UseGuards(AuthGuard("jwt"))
  getMe(@Req() req: any) {
    return this.authService.getProfile(req.user.id);
  }

  @Patch("profile")
  @UseGuards(AuthGuard("jwt"))
  updateProfile(@Req() req: any, @Body() dto: UpdateProfileDto) {
    return this.authService.updateProfile(req.user.id, dto);
  }

  @Patch("password")
  @UseGuards(AuthGuard("jwt"))
  changePassword(@Req() req: any, @Body() body: { currentPassword: string; newPassword: string }) {
    return this.authService.changePassword(req.user.id, body.currentPassword, body.newPassword);
  }

  @Post("delete-account")
  @UseGuards(AuthGuard("jwt"))
  deleteAccount(@Req() req: any, @Body() body: { reason?: string }) {
    return this.authService.requestAccountDeletion(req.user.id, body.reason);
  }
}
