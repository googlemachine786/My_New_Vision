import { Controller, Get, Post, Body, Param, Sse, Query, UseGuards, Req } from "@nestjs/common";
import { AuthGuard } from "@nestjs/passport";
import { ChatService } from "./chat.service";
import { Observable } from "rxjs";

@Controller("chat")
@UseGuards(AuthGuard("jwt"))
export class ChatController {
  constructor(private chatService: ChatService) {}

  @Post("conversations")
  createConversation(@Req() req: any) {
    return this.chatService.createConversation(req.user.id);
  }

  @Get("conversations")
  getConversations(@Req() req: any) {
    return this.chatService.getConversations(req.user.id);
  }

  @Get("conversations/:id/messages")
  getMessages(@Param("id") id: string, @Req() req: any) {
    return this.chatService.getMessages(id, req.user.id);
  }

  @Sse("stream")
  stream(
    @Query("prompt") prompt: string,
    @Query("conversationId") conversationId: string,
    @Req() req: any
  ): Observable<MessageEvent> {
    return this.chatService.streamResponse(prompt, conversationId, req.user.id);
  }
}
