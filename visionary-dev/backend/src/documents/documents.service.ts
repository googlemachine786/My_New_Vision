import { Injectable } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import { PrismaService } from "../prisma/prisma.service";
import { createClient } from "@supabase/supabase-js";

@Injectable()
export class DocumentsService {
  private supabase;

  constructor(private prisma: PrismaService, private config: ConfigService) {
    this.supabase = createClient(
      this.config.get("SUPABASE_URL"),
      this.config.get("SUPABASE_SERVICE_KEY")
    );
  }

  async upload(file: any, userId: string) {
    const filename = `${userId}/${Date.now()}-${file.originalname}`;

    // Upload to Supabase Storage
    const { error } = await this.supabase.storage
      .from(this.config.get("SUPABASE_STORAGE_BUCKET") || "uploads")
      .upload(filename, file.buffer, { contentType: file.mimetype });

    if (error) throw error;

    const { data: urlData } = this.supabase.storage
      .from("uploads")
      .getPublicUrl(filename);

    const document = await this.prisma.document.create({
      data: {
        userId,
        filename: file.originalname,
        storageUrl: urlData.publicUrl,
        status: "PROCESSING",
      },
    });

    // TODO: Queue PDF processing job (Phase 5)
    // this.queue.add('process-pdf', { documentId: document.id });

    return document;
  }

  async list(userId: string) {
    return this.prisma.document.findMany({
      where: { userId },
      orderBy: { createdAt: "desc" },
    });
  }

  async get(id: string, userId: string) {
    return this.prisma.document.findFirstOrThrow({ where: { id, userId } });
  }
}
