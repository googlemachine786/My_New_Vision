import { Module } from "@nestjs/common";
import { OrchestratorService } from "./orchestrator.service";
import { ContentAgentService } from "../agents/content/content-agent.service";
import { DiagramAgentService } from "../agents/diagram/diagram-agent.service";
import { ResourceAgentService } from "../agents/resource/resource-agent.service";

@Module({
  providers: [
    OrchestratorService,
    ContentAgentService,
    DiagramAgentService,
    ResourceAgentService,
  ],
  exports: [OrchestratorService],
})
export class OrchestratorModule {}
