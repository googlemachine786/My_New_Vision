import { Injectable, Logger } from "@nestjs/common";
import { HealthCheckError, HealthIndicator, HealthIndicatorResult } from "@nestjs/terminus";
import { PrismaService } from "../prisma/prisma.service";

@Injectable()
export class PrismaHealthIndicator extends HealthIndicator {
  private readonly logger = new Logger(PrismaHealthIndicator.name);

  constructor(private prisma: PrismaService) {
    super();
  }

  async isHealthy(key: string): Promise<HealthIndicatorResult> {
    try {
      await this.prisma.$queryRaw`SELECT 1`;
      return this.getStatus(key, true);
    } catch (e) {
      this.logger.error(`Prisma check failed: ${(e as Error).message}`);
      throw new HealthCheckError("Prisma check failed", this.getStatus(key, false));
    }
  }
}
