import { Controller, Post, Get, Body, UseGuards, Req, Res } from "@nestjs/common";
import type { Response } from "express";
import { AuthGuard } from "@nestjs/passport";
import { OnboardingService } from "./onboarding.service";
import { OnboardingCompleteDto } from "./dto/onboarding-complete.dto";
import { setAuthCookie } from "../auth/auth.cookie";

@Controller("onboarding")
@UseGuards(AuthGuard("jwt"))
export class OnboardingController {
  constructor(private onboardingService: OnboardingService) {}

  @Get("status")
  getStatus(@Req() req: any) {
    return this.onboardingService.getStatus(req.user.id);
  }

  @Post("complete")
  async complete(
    @Body() dto: OnboardingCompleteDto,
    @Req() req: any,
    @Res({ passthrough: true }) res: Response
  ) {
    const result = await this.onboardingService.complete(req.user.id, dto);
    setAuthCookie(res, result.accessToken);
    return result;
  }
}
