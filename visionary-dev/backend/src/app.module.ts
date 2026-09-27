import { Module } from "@nestjs/common";
import { ConfigModule } from "@nestjs/config";
import { AuthModule } from "./auth/auth.module";
import { ChatModule } from "./chat/chat.module";
import { DocumentsModule } from "./documents/documents.module";
import { JobsModule } from "./jobs/jobs.module";
import { OnboardingModule } from "./onboarding/onboarding.module";
import { BoardsModule } from "./boards/boards.module";
import { GradesModule } from "./grades/grades.module";
import { SubjectsModule } from "./subjects/subjects.module";
import { PrismaModule } from "./prisma/prisma.module";
import { RedisModule } from "./redis/redis.module";
import { HealthModule } from "./health/health.module";

@Module({
  imports: [
    ConfigModule.forRoot({ isGlobal: true }),
    PrismaModule,
    RedisModule,
    AuthModule,
    OnboardingModule,
    BoardsModule,
    GradesModule,
    SubjectsModule,
    ChatModule,
    DocumentsModule,
    JobsModule,
    HealthModule,
  ],
})
export class AppModule {}
