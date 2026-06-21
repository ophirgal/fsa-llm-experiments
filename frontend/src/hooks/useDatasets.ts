import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import api from '../api/client'

interface Dataset {
  id: number
  name: string
  created_at: string
}

export function useDatasets() {
  return useQuery({
    queryKey: ['datasets'],
    queryFn: () => api.get<Dataset[]>('/datasets').then(res => res.data),
  })
}

export function useCreateDataset() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ name, file }: { name: string; file: File }) => {
      const form = new FormData()
      form.append('name', name)
      form.append('file', file)
      return api.post('/datasets', form)
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['datasets'] }),
  })
}
