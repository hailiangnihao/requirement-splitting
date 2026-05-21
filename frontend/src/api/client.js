const API_BASE = import.meta.env.VITE_API_BASE || 'http://localhost:8080';

async function request(path, options = {}) {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  });

  const text = await response.text();
  const data = text ? JSON.parse(text) : null;
  if (!response.ok) {
    throw new Error(data?.error || `请求失败：${response.status}`);
  }
  return data;
}

export const api = {
  listProjects: () => request('/api/projects'),
  createProject: (payload) => request('/api/projects', { method: 'POST', body: JSON.stringify(payload) }),
  getProject: (id) => request(`/api/projects/${id}`),
  listRequirements: (projectId) => request(`/api/projects/${projectId}/requirements`),
  createRequirement: (projectId, payload) => request(`/api/projects/${projectId}/requirements`, { method: 'POST', body: JSON.stringify(payload) }),
  splitRequirement: (projectId, payload) => request(`/api/projects/${projectId}/ai/split-requirement`, { method: 'POST', body: JSON.stringify(payload) }),
  listDrafts: (projectId) => request(`/api/projects/${projectId}/ai-drafts`),
  publishDraft: (projectId, draftId) => request(`/api/projects/${projectId}/ai-drafts/${draftId}/publish`, { method: 'POST' }),
  getPlan: (projectId) => request(`/api/projects/${projectId}/plan`),
  listDevTasks: (projectId) => request(`/api/projects/${projectId}/dev-tasks`),
  updateDevTaskStatus: (projectId, taskId, status) => request(`/api/projects/${projectId}/dev-tasks/${taskId}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  listTestCases: (projectId) => request(`/api/projects/${projectId}/test-cases`),
  confirmTestCase: (projectId, testCaseId) => request(`/api/projects/${projectId}/test-cases/${testCaseId}/confirm`, { method: 'POST' }),
  runAiTest: (projectId, testCaseId) => request(`/api/projects/${projectId}/test-cases/${testCaseId}/ai-run`, { method: 'POST' }),
  listTestRuns: (projectId) => request(`/api/projects/${projectId}/test-runs`),
  reviewTestRun: (projectId, testRunId, status) => request(`/api/projects/${projectId}/test-runs/${testRunId}/review`, { method: 'POST', body: JSON.stringify({ status }) }),
  listDefects: (projectId) => request(`/api/projects/${projectId}/defects`),
  createDefect: (projectId, payload) => request(`/api/projects/${projectId}/defects`, { method: 'POST', body: JSON.stringify(payload) }),
  updateDefectStatus: (projectId, defectId, status) => request(`/api/projects/${projectId}/defects/${defectId}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  listChanges: (projectId) => request(`/api/projects/${projectId}/changes`),
  createChange: (projectId, payload) => request(`/api/projects/${projectId}/changes`, { method: 'POST', body: JSON.stringify(payload) }),
  analyzeChange: (projectId, changeId) => request(`/api/projects/${projectId}/changes/${changeId}/analyze`, { method: 'POST' }),
  updateChangeStatus: (projectId, changeId, status) => request(`/api/projects/${projectId}/changes/${changeId}/status`, { method: 'PATCH', body: JSON.stringify({ status }) }),
  getHealth: (projectId) => request(`/api/projects/${projectId}/health`)
};
