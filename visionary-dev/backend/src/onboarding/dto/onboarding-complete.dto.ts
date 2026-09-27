import { IsIn, IsOptional, IsString, MaxLength } from "class-validator";

export class OnboardingCompleteDto {
  @IsIn(["student", "teacher", "organization"])
  category!: "student" | "teacher" | "organization";

  @IsOptional()
  @IsString()
  @MaxLength(120)
  fullName?: string;

  @IsOptional()
  @IsString()
  @MaxLength(40)
  grade?: string;

  @IsOptional()
  @IsString()
  @MaxLength(40)
  board?: string;

  @IsOptional()
  @IsString()
  @MaxLength(40)
  subject?: string;

  @IsOptional()
  @IsString()
  @MaxLength(120)
  organizationName?: string;

  @IsOptional()
  @IsString()
  @MaxLength(60)
  organizationType?: string;
}
