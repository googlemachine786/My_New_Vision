import { Injectable } from "@nestjs/common";
import { PrismaService } from "../prisma/prisma.service";

@Injectable()
export class GradesService {
  constructor(private prisma: PrismaService) {}

  async getAll() {
    return this.prisma.gradeMaster.findMany({
      where: { isActive: true },
      orderBy: { sortOrder: "asc" },
      select: { id: true, name: true, sortOrder: true },
    });
  }
}
