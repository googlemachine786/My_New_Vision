import { Injectable } from "@nestjs/common";
import { PrismaService } from "../prisma/prisma.service";
import { OrchestratorService } from "../orchestrator/orchestrator.service";
import { Observable, Subscriber } from "rxjs";

@Injectable()
export class ChatService {
  constructor(
    private prisma: PrismaService,
    private orchestrator: OrchestratorService
  ) {}

  async createConversation(userId: string) {
    return this.prisma.conversation.create({
      data: { userId },
    });
  }

  async getConversations(userId: string) {
    return this.prisma.conversation.findMany({
      where: { userId },
      orderBy: { updatedAt: "desc" },
      take: 50,
    });
  }

  async getMessages(conversationId: string, userId: string) {
    // Verify ownership
    await this.prisma.conversation.findFirstOrThrow({
      where: { id: conversationId, userId },
    });

    return this.prisma.message.findMany({
      where: { conversationId },
      orderBy: { createdAt: "asc" },
    });
  }

  streamResponse(
    prompt: string,
    conversationId: string,
    userId: string
  ): Observable<MessageEvent> {
    return new Observable((subscriber) => {
      this.handleStream(prompt, conversationId, userId, subscriber);
    });
  }

  private async handleStream(
    prompt: string,
    conversationId: string,
    userId: string,
    subscriber: Subscriber<MessageEvent>
  ) {
    let conversation: { id: string } | null;
    try {
      conversation = await this.prisma.conversation.findFirst({
        where: { id: conversationId, userId },
        select: { id: true },
      });
    } catch {
      subscriber.next({ data: JSON.stringify({ error: "Forbidden" }) } as MessageEvent);
      subscriber.complete();
      return;
    }
    if (!conversation) {
      subscriber.next({ data: JSON.stringify({ error: "Forbidden" }) } as MessageEvent);
      subscriber.complete();
      return;
    }

    try {
      await this.prisma.message.create({
        data: { conversationId, role: "user", content: prompt },
      });

      // Stream via orchestrator
      let fullContent = "";
      await this.orchestrator.streamResponse(prompt, (chunk: string) => {
        fullContent += chunk;
        subscriber.next({ data: JSON.stringify({ token: chunk }) } as MessageEvent);
      });

      // Save assistant message
      await this.prisma.message.create({
        data: { conversationId, role: "assistant", content: fullContent },
      });

      subscriber.next({ data: JSON.stringify({ done: true }) } as MessageEvent);
      subscriber.complete();
    } catch (err) {
      subscriber.next({ data: JSON.stringify({ error: "Stream failed" }) } as MessageEvent);
      subscriber.complete();
    }
  }
}
