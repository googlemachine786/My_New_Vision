import {
  Controller,
  Post,
  Get,
  Param,
  UseGuards,
  Req,
  UseInterceptors,
  UploadedFile,
} from "@nestjs/common";
import { AuthGuard } from "@nestjs/passport";
import { FileInterceptor } from "@nestjs/platform-express";
import { DocumentsService } from "./documents.service";

@Controller("documents")
@UseGuards(AuthGuard("jwt"))
export class DocumentsController {
  constructor(private documentsService: DocumentsService) {}

  @Post("upload")
  @UseInterceptors(FileInterceptor("file"))
  upload(@UploadedFile() file: any, @Req() req: any) {
    return this.documentsService.upload(file, req.user.id);
  }

  @Get()
  list(@Req() req: any) {
    return this.documentsService.list(req.user.id);
  }

  @Get(":id")
  get(@Param("id") id: string, @Req() req: any) {
    return this.documentsService.get(id, req.user.id);
  }
}
