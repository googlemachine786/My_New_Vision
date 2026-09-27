import { Injectable } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import OpenAI from "openai";
import { ContentAgentService } from "../agents/content/content-agent.service";
import { DiagramAgentService } from "../agents/diagram/diagram-agent.service";
import { ResourceAgentService } from "../agents/resource/resource-agent.service";

interface Intent {
  explanation: boolean;
  diagram: boolean;
  image: boolean;
  resource: boolean;
}

@Injectable()
export class OrchestratorService {
  private openai: OpenAI;

  constructor(
    private config: ConfigService,
    private contentAgent: ContentAgentService,
    private diagramAgent: DiagramAgentService,
    private resourceAgent: ResourceAgentService
  ) {
    this.openai = new OpenAI({ apiKey: this.config.get("OPENAI_API_KEY") });
  }

  async detectIntent(prompt: string): Promise<Intent> {
    // Fast intent classification
    const result = await this.openai.chat.completions.create({
      model: "gpt-4o-mini",
      messages: [
        {
          role: "system",
          content: `Classify this educational prompt. Return JSON only: { "explanation": bool, "diagram": bool, "image": bool, "resource": bool }`,
        },
        { role: "user", content: prompt },
      ],
      response_format: { type: "json_object" },
      max_tokens: 100,
    });

    try {
      return JSON.parse(result.choices[0].message.content || "{}");
    } catch {
      return { explanation: true, diagram: false, image: false, resource: false };
    }
  }

  async streamResponse(prompt: string, onToken: (chunk: string) => void): Promise<void> {
    const intent = await this.detectIntent(prompt);

    // Always stream explanation via content agent
    await this.contentAgent.stream(prompt, onToken);

    // Fire non-streaming agents if needed (results sent as separate SSE events)
    const tasks: Promise<void>[] = [];
    if (intent.diagram) tasks.push(this.diagramAgent.generate(prompt, onToken));
    if (intent.resource) tasks.push(this.resourceAgent.fetch(prompt, onToken));

    await Promise.allSettled(tasks);
  }
}
