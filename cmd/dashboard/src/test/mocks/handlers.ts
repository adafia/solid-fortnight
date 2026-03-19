import { http, HttpResponse } from 'msw'
import { Project } from '../../types/api'

const API_BASE = 'http://127.0.0.1:8080/api/v1/management'

// Centralized mock data
export const mockProjects: Project[] = [
  {
    id: 'proj-123',
    name: 'Mock Project',
    description: 'A project from MSW',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }
]

export const handlers = [
  // List Projects
  http.get(`${API_BASE}/projects/`, () => {
    return HttpResponse.json(mockProjects)
  }),

  // Create Project
  http.post(`${API_BASE}/projects/`, async ({ request }) => {
    const newProject = await request.json() as any
    const project: Project = {
      id: `proj-${Math.random().toString(36).substr(2, 9)}`,
      name: newProject.name,
      description: newProject.description,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString()
    }
    return HttpResponse.json(project, { status: 201 })
  }),

  // Get Project Details
  http.get(`${API_BASE}/projects/:id`, ({ params }) => {
    const project = mockProjects.find(p => p.id === params.id) || mockProjects[0]
    return HttpResponse.json(project)
  }),
]
