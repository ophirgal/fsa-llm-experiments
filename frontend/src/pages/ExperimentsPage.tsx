import { useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { useDatasets } from '@/hooks/useDatasets'
import {
  useCreateExperiment,
  useExperiments,
  useStartExperiment,
  type CreateExperimentInput,
} from '@/hooks/useExperiments'

const STATUS_CLASS: Record<string, string> = {
  ready: 'text-muted-foreground',
  'in progress': 'text-blue-600',
  done: 'text-green-600',
  failed: 'text-destructive',
}

function CreateExperimentDialog() {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [datasetId, setDatasetId] = useState('')
  const [prompts, setPrompts] = useState<string[]>([''])
  const [judgePrompt, setJudgePrompt] = useState('')

  const { data: datasets } = useDatasets()
  const { mutate, isPending, error } = useCreateExperiment()

  function reset() {
    setName('')
    setDatasetId('')
    setPrompts([''])
    setJudgePrompt('')
  }

  const isValid = name.trim() && datasetId && prompts.every(p => p.trim()) && judgePrompt.trim()

  function setPrompt(index: number, value: string) {
    setPrompts(prev => prev.map((p, i) => (i === index ? value : p)))
  }

  function addPrompt() {
    if (prompts.length < 3) setPrompts(prev => [...prev, ''])
  }

  function removePrompt(index: number) {
    setPrompts(prev => prev.filter((_, i) => i !== index))
  }

  function buildInput(): CreateExperimentInput {
    return {
      name: name.trim(),
      dataset_id: Number(datasetId),
      prompts: prompts.map(p => p.trim()),
      judge_prompt: judgePrompt.trim(),
    }
  }

  function handleCreate() {
    if (!isValid) return
    mutate({ input: buildInput(), start: false }, { onSuccess: () => { setOpen(false); reset() } })
  }

  function handleCreateAndRun() {
    if (!isValid) return
    mutate({ input: buildInput(), start: true }, { onSuccess: () => { setOpen(false); reset() } })
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { setOpen(o); if (!o) reset() }}>
      <DialogTrigger asChild>
        <Button>Create Experiment</Button>
      </DialogTrigger>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Create Experiment</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div className="space-y-1">
            <label className="text-sm font-medium">Name</label>
            <Input
              placeholder="My experiment"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Dataset</label>
            <select
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              value={datasetId}
              onChange={(e) => setDatasetId(e.target.value)}
            >
              <option value="">Select a dataset…</option>
              {datasets?.map((d) => (
                <option key={d.id} value={d.id}>{d.name}</option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">Prompt Variants</label>
            {prompts.map((p, i) => (
              <div key={i} className="relative">
                <textarea
                  className="flex min-h-[90px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring resize-none"
                  placeholder={`Variant ${i + 1}: You are a helpful assistant…`}
                  value={p}
                  onChange={(e) => setPrompt(i, e.target.value)}
                />
                {prompts.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removePrompt(i)}
                    className="absolute top-1.5 right-1.5 rounded p-0.5 text-muted-foreground hover:text-foreground hover:bg-muted"
                    aria-label="Remove variant"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                )}
              </div>
            ))}
            {prompts.length < 3 && (
              <Button type="button" variant="outline" size="sm" onClick={addPrompt}>
                + Add Prompt Variant
              </Button>
            )}
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Judge Prompt</label>
            <textarea
              className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring resize-none"
              placeholder="Grade the response on a scale of 0–1…"
              value={judgePrompt}
              onChange={(e) => setJudgePrompt(e.target.value)}
            />
          </div>
          {error && (
            <p className="text-sm text-destructive">
              {(error as { response?: { data?: { error?: string } } })?.response?.data?.error ?? 'Failed to create experiment.'}
            </p>
          )}
        </div>
        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={handleCreate} disabled={isPending || !isValid}>
            Create
          </Button>
          <Button onClick={handleCreateAndRun} disabled={isPending || !isValid}>
            {isPending ? 'Creating…' : 'Create & Run'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

function RunButton({ id }: { id: number }) {
  const { mutate, isPending } = useStartExperiment()
  return (
    <Button size="sm" variant="outline" disabled={isPending} onClick={() => mutate(id)}>
      {isPending ? 'Starting…' : 'Run'}
    </Button>
  )
}

export default function ExperimentsPage() {
  const { data: experiments, isLoading, isError } = useExperiments()

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold">Experiments</h1>
        <CreateExperimentDialog />
      </div>

      {isLoading && <p className="text-muted-foreground">Loading…</p>}
      {isError && <p className="text-destructive">Failed to load experiments.</p>}
      {!isLoading && !isError && experiments && experiments.length === 0 && (
        <p className="text-muted-foreground">No experiments yet.</p>
      )}
      {!isLoading && !isError && experiments && experiments.length > 0 && (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-16">ID</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Dataset</TableHead>
              <TableHead>Score</TableHead>
              <TableHead>Started</TableHead>
              <TableHead>Ended</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {experiments.map((e) => (
              <TableRow key={e.id}>
                <TableCell>{e.id}</TableCell>
                <TableCell>{e.name}</TableCell>
                <TableCell>
                  <span className={cn('font-medium', STATUS_CLASS[e.status] ?? '')}>
                    {e.status}
                  </span>
                </TableCell>
                <TableCell>{e.dataset_id}</TableCell>
                <TableCell>{e.total_score != null ? e.total_score : '—'}</TableCell>
                <TableCell>{e.start_time ? new Date(e.start_time).toLocaleString() : '—'}</TableCell>
                <TableCell>{e.end_time ? new Date(e.end_time).toLocaleString() : '—'}</TableCell>
                <TableCell>
                  {e.status === 'ready' && <RunButton id={e.id} />}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  )
}
