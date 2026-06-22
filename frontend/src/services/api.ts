const API_BASE_URL = import.meta.env.QCLI_API_BASE_URL || 'http://localhost:8080';

export interface HealthResponse {
  status: string;
  api: string;
  database: string;
  error?: string;
}

export async function getHealth(): Promise<HealthResponse> {
  const response = await fetch(`${API_BASE_URL}/api/v1/health`);
  const payload = (await response.json()) as HealthResponse;

  if (!response.ok) {
    throw new Error(payload.error || 'Backend health check failed.');
  }

  return payload;
}
