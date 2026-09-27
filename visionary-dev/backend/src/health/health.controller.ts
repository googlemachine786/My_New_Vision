import { Controller, Get } from "@nestjs/common";
import { HealthCheck, HealthCheckService } from "@nestjs/terminus";
import { PrismaHealthIndicator } from "./prisma.health";
import { RedisHealthIndicator } from "./redis.health";

@Controller("health")
export class HealthController {
  constructor(
    private health: HealthCheckService,
    private prismaIndicator: PrismaHealthIndicator,
    private redisIndicator: RedisHealthIndicator,
  ) {}

  // Liveness — process is alive. No external deps; always 200 if Nest is responding.
  @Get("live")
  @HealthCheck()
  liveness() {
    return this.health.check([]);
  }

  // Readiness — dependencies (Postgres, Redis) are healthy. 200 if all up, 503 if any down.
  @Get("ready")
  @HealthCheck()
  readiness() {
    return this.health.check([
      () => this.prismaIndicator.isHealthy("db"),
      () => this.redisIndicator.isHealthy("redis"),
    ]);
  }

  // Default — same as readiness; used by Docker HEALTHCHECK.
  @Get()
  @HealthCheck()
  check() {
    return this.health.check([
      () => this.prismaIndicator.isHealthy("db"),
      () => this.redisIndicator.isHealthy("redis"),
    ]);
  }
}
