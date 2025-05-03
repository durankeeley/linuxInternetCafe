'use client'

import { useEffect, useState } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Dialog, DialogTrigger, DialogContent } from '@/components/ui/dialog'

const formatTimeLeft = (seconds: number) => {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

type Computer = {
  id: number
  hostname: string
  ip: string
  ssh_port: number
  ssh_user: string
  assigned: string
  status: string
  session_expires_at?: string
}

export default function Dashboard() {
  const [computers, setComputers] = useState<Computer[]>([])
  const [timeLeft, setTimeLeft] = useState<{ [key: number]: number }>({})
  const [startModalOpen, setStartModalOpen] = useState<number | null>(null)
  const [extendModalOpen, setExtendModalOpen] = useState<number | null>(null)
  const [duration, setDuration] = useState<number>(15)

  const fetchComputers = async () => {
    const token = localStorage.getItem('token')
    const res = await fetch('/api/computers', {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
    if (res.ok) {
      const data: Computer[] = await res.json()
      setComputers(data)

      const now = Date.now()
      const newTimeLeft: { [key: number]: number } = {}
      data.forEach(c => {
        if (c.session_expires_at) {
          const expires = new Date(c.session_expires_at).getTime()
          const diff = Math.max(0, Math.floor((expires - now) / 1000))
          newTimeLeft[c.id] = diff
        }
      })
      setTimeLeft(newTimeLeft)
    }
  }

  useEffect(() => {
    fetchComputers()
    const interval = setInterval(fetchComputers, 5000)
    return () => clearInterval(interval)
  }, [])

  useEffect(() => {
    const countdown = setInterval(() => {
      setTimeLeft(prev => {
        const updated: { [key: number]: number } = {}
        Object.keys(prev).forEach(id => {
          const remaining = prev[+id] - 1
          updated[+id] = Math.max(0, remaining)
        })
        return updated
      })
    }, 1000)
    return () => clearInterval(countdown)
  }, [])

  const handleSession = async (id: number, type: 'start' | 'extend') => {
    const token = localStorage.getItem('token')
    const res = await fetch(`/api/session/${type}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ computer_id: id, minutes: duration }),
    })
    if (res.ok) {
      if (type === 'start') setStartModalOpen(null)
      else setExtendModalOpen(null)
      fetchComputers()
    }
  }

  const handleEndSession = async (id: number) => {
    const token = localStorage.getItem('token')
    await fetch(`/api/session/end`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ computer_id: id }),
    })
    fetchComputers()
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-4">
      {computers.map(c => (
        <Card key={c.id} className="p-4 space-y-2">
          <h3 className="text-xl font-bold">{c.hostname}</h3>
          <p>Status: {c.status}</p>
          <p>Assigned: {c.assigned}</p>
          {timeLeft[c.id] !== undefined && (
            <p>Time Left: {formatTimeLeft(timeLeft[c.id])}</p>
          )}
          <div className="space-x-2">
            <Button onClick={() => handleEndSession(c.id)}>End</Button>
            <Dialog open={startModalOpen === c.id} onOpenChange={open => setStartModalOpen(open ? c.id : null)}>
              <DialogTrigger asChild>
                <Button>Start</Button>
              </DialogTrigger>
              <DialogContent>
                <h4 className="text-lg font-semibold mb-2">Start Session</h4>
                <input
                  type="number"
                  className="border px-2 py-1 w-full mb-2"
                  value={duration}
                  onChange={e => setDuration(+e.target.value)}
                />
                <Button onClick={() => handleSession(c.id, 'start')}>Confirm</Button>
              </DialogContent>
            </Dialog>
            <Dialog open={extendModalOpen === c.id} onOpenChange={open => setExtendModalOpen(open ? c.id : null)}>
              <DialogTrigger asChild>
                <Button>Extend</Button>
              </DialogTrigger>
              <DialogContent>
                <h4 className="text-lg font-semibold mb-2">Extend Session</h4>
                <input
                  type="number"
                  className="border px-2 py-1 w-full mb-2"
                  value={duration}
                  onChange={e => setDuration(+e.target.value)}
                />
                <Button onClick={() => handleSession(c.id, 'extend')}>Confirm</Button>
              </DialogContent>
            </Dialog>
          </div>
        </Card>
      ))}
    </div>
  )
}
