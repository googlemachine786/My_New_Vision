import { IsNotEmpty, IsString } from "class-validator";

export class CreateImageJobDto {
  @IsString()
  @IsNotEmpty()
  prompt!: string;
}
