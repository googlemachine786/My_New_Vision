import { Injectable } from "@nestjs/common";
import { ConfigService } from "@nestjs/config";
import OpenAI from "openai";

@Injectable()
export class ResourceAgentService {
  private openai: OpenAI;

  constructor(private config: ConfigService) {
    this.openai = new OpenAI({ apiKey: this.config.get("OPENAI_API_KEY") });
  }

  async fetch(prompt: string, onToken: (chunk: string) => void): Promise<void> {
    // TODO: Integrate YouTube Data API v3 for real video suggestions
    // For now, return placeholder resources
    const resources = [
      {
        title: `Learn more about ${prompt.split(" ").slice(0, 3).join(" ")}`,
        url: "https://www.youtube.com",
      },
    ];

    onToken(`\n\n[RESOURCES]${JSON.stringify(resources)}[/RESOURCES]`);
  }
}
