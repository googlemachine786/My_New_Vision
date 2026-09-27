import { Controller, Get } from "@nestjs/common";
import { BoardsService } from "./boards.service";

@Controller("boards")
export class BoardsController {
  constructor(private boardsService: BoardsService) {}

  // Public endpoint — no auth required
  @Get()
  getAll() {
    return this.boardsService.getAll();
  }
}
