import { useEffect, useState } from 'react'

export function Avatar({ seed, avatar, size }: { seed: string; avatar?: string; size: number }) {
  const [failed, setFailed] = useState(false)
  useEffect(() => setFailed(false), [avatar])

  let url = ''
  try {
    const u = new URL(avatar || '')
    if (u.protocol === 'http:' || u.protocol === 'https:') url = u.toString()
  } catch {}

  // deterministic pastel color from the name, same formula as the legacy client
  let a = 0,
    b = 0,
    c = 0
  for (let i = 0; i < seed.length; i++) {
    a += seed.charCodeAt(i) * 3
    b += seed.charCodeAt(i) * 5
    c += seed.charCodeAt(i) * 7
  }
  const bg = `rgb(${128 + (a % 128)},${128 + (b % 128)},${128 + (c % 128)})`

  return (
    <div
      className="inline-block overflow-hidden rounded-full bg-white text-center text-black select-none"
      style={{ width: size, height: size, lineHeight: size + 'px', fontSize: size / 2, backgroundColor: bg }}
    >
      {url && !failed ? (
        <img src={url} alt={seed} className="block h-full w-full object-cover" onError={() => setFailed(true)} />
      ) : (
        seed.substring(0, 2).toUpperCase()
      )}
    </div>
  )
}
