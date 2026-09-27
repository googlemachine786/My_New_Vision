import { Controller, Post, Get, Body, Param, UseGuards, Req } from "@nestjs/common";
import { AuthGuard } from "@nestjs/passport";
import { JobsService } from "./jobs.service";
import { CreateImageJobDto } from "./dto/create-image-job.dto";

@Controller("jobs")
@UseGuards(AuthGuard("jwt"))
export class JobsController {
  constructor(private jobsService: JobsService) {}

  @Post("image")
  createImageJob(@Body() dto: CreateImageJobDto, @Req() req: any) {
    return this.jobsService.createImageJob(dto.prompt, req.user.id);
  }

  @Get(":id")
  getJob(@Param("id") id: string, @Req() req: any) {
    return this.jobsService.getJob(id, req.user.id);
  }
}
