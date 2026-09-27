import { Controller, Get } from "@nestjs/common";
import { GradesService } from "./grades.service";

@Controller("grades")
export class GradesController {
  constructor(private gradesService: GradesService) {}

  // Public endpoint — no auth required
  @Get()
  getAll() {
    return this.gradesService.getAll();
  }
}
