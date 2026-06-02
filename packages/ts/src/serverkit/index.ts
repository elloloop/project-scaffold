import { displayName } from "../platform/index";

export type HealthStatus = {
  service: string;
  status: "ok";
};

export function healthPayload(service: string): HealthStatus {
  return {
    service: displayName(service),
    status: "ok",
  };
}
