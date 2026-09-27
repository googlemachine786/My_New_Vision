import { Injectable, Logger } from "@nestjs/common";
import { HealthCheckError, HealthIndicator, HealthIndicatorResult } from "@nestjs/terminus";
import { RedisService } from "../redis/redis.service";

@Injectable()
export class RedisHealthIndicator extends HealthIndicator {
  private readonly logger = new Logger(RedisHealthIndicator.name);

  constructor(private redisService: RedisService) {
    super();
  }

  async isHealthy(key: string): Promise<HealthIndicatorResult> {
    try {
      const pong = await this.redisService.redis.ping();
      if (pong !== "PONG") throw new Error(`Unexpected ping response: ${pong}`);
      return this.getStatus(key, true);
    } catch (e) {
      this.logger.error(`Redis check failed: ${(e as Error).message}`);
      throw new HealthCheckError("Redis check failed", this.getStatus(key, false));
    }
  }
}
