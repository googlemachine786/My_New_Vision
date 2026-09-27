import { Injectable } from "@nestjs/common";
import { PrismaService } from "../prisma/prisma.service";

@Injectable()
export class JobsService {
  constructor(private prisma: PrismaService) {}

  async createImageJob(prompt: string, userId: string) {
    const job = await this.prisma.job.create({
      data: { userId, type: "IMAGE_GENERATION", status: "PENDING", prompt },
    });

    // TODO: Add to BullMQ queue (Phase 6)
    // await this.queue.add('generate-image', { jobId: job.id, prompt });

    return { jobId: job.id, status: job.status };
  }

  async getJob(id: string, userId: string) {
    return this.prisma.job.findFirstOrThrow({ where: { id, userId } });
  }
}
