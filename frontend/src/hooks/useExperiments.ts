import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import api from '../api/client'

export interface Experiment {
  id: number
  name: string
  status: 'ready' | 'in progress' | 'done' | 'failed'
  dataset_id: number
  total_score: number | null
  start_time: string | null
  end_time: string | null
}

export interface CreateExperimentInput {
  name: string
  dataset_id: number
  prompts: string[]
  judge_prompt: string
}

export function useExperiments() {
  return useQuery({
    queryKey: ['experiments'],
    queryFn: () => api.get<Experiment[]>('/experiments').then(res => res.data),
    refetchInterval: (query) => {
      const hasInProgress = query.state.data?.some(e => e.status === 'in progress')
      return hasInProgress ? 3000 : false
    },
  })
}

export function useCreateExperiment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ input, start }: { input: CreateExperimentInput; start: boolean }) =>
      api.post('/experiments' + (start ? '?isStarted=true' : ''), input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['experiments'] }),
  })
}

export function useStartExperiment() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.put(`/experiments/${id}`, { status: 'in progress' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['experiments'] }),
  })
}
