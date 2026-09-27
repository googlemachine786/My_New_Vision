import { Controller, Get, Post, Param, UseGuards, Request, Res } from "@nestjs/common";
import type { Response } from "express";
import { AuthGuard } from "@nestjs/passport";
import { SubjectsService } from "./subjects.service";
import { setAuthCookie } from "../auth/auth.cookie";

@Controller("subjects")
@UseGuards(AuthGuard("jwt"))
export class SubjectsController {
  constructor(private subjectsService: SubjectsService) {}

  @Get("me")
  getMySubjects(@Request() req: any) {
    return this.subjectsService.getForUser(req.user.id);
  }

  @Get("started")
  getStarted(@Request() req: any) {
    return this.subjectsService.getStartedSubject(req.user.id);
  }

  @Get("my-tracks")
  getAllStartedChapters(@Request() req: any) {
    return this.subjectsService.getAllStartedChapters(req.user.id);
  }

  @Post(":id/start")
  async startSubject(
    @Request() req: any,
    @Param("id") subjectId: string,
    @Res({ passthrough: true }) res: Response
  ) {
    const result = await this.subjectsService.startSubject(req.user.id, subjectId);
    setAuthCookie(res, result.accessToken);
    return result;
  }

  @Get("chapter-stats")
  getChapterStats(@Request() req: any) {
    return this.subjectsService.getChapterStats(req.user.id);
  }

  @Get("all-tracked-chapters")
  getAllTrackedChapters(@Request() req: any) {
    return this.subjectsService.getAllTrackedChapters(req.user.id);
  }

  @Get("recent-chapters")
  getRecentChapters(@Request() req: any) {
    return this.subjectsService.getRecentChapters(req.user.id);
  }

  @Get("chapters/:chapterId/topics")
  getChapterTopics(@Param("chapterId") chapterId: string) {
    return this.subjectsService.getChapterTopics(chapterId);
  }

  @Post("chapters/:chapterId/start")
  async startChapter(
    @Request() req: any,
    @Param("chapterId") chapterId: string,
    @Res({ passthrough: true }) res: Response
  ) {
    const result = await this.subjectsService.startChapter(req.user.id, chapterId);
    setAuthCookie(res, result.accessToken);
    return result;
  }

  @Get(":id/chapters")
  getChapters(@Request() req: any, @Param("id") subjectId: string) {
    return this.subjectsService.getChapters(req.user.id, subjectId);
  }
}
