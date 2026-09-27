import { IsEmail, IsString, Matches } from "class-validator";

export class VerifyOtpDto {
  @IsEmail()
  email!: string;

  @IsString()
  @Matches(/^\d{6}$/, { message: "otp must be a 6 digit numeric code" })
  otp!: string;
}
