import { Injectable } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import OpenAI from "openai";

@Injectable()
export class DiagramAgentService {
  private openai: OpenAI;

  constructor(private config: ConfigService) {
    this.openai = new OpenAI({ apiKey: this.config.get("OPENAI_API_KEY") });
  }

  async generate(prompt: string, onToken: (chunk: string) => void): Promise<void> {
    const result = await this.openai.chat.completions.create({
      model: "gpt-4o-mini",
      messages: [
        {
          role: "system",
          content: "Generate a Mermaid diagram for the concept. Return ONLY the mermaid code block, no extra text. Start with ```mermaid",
        },
        { role: "user", content: `Create a diagram for: ${prompt}` },
      ],
    });

    const diagram = result.choices[0].message.content || "";
    // Send diagram as a special SSE event
    onToken(`\n\n[DIAGRAM]${diagram}[/DIAGRAM]`);
  }
}
