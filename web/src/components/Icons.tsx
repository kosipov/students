// Lucide-style icons at stroke 1.5, as the Industry system asks.

export function SearchIcon() {
  return (
    <svg className="icn" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-3.5-3.5" />
    </svg>
  )
}

export function PlusIcon() {
  return (
    <svg className="icn" viewBox="0 0 24 24" aria-hidden="true">
      <path d="M12 5v14M5 12h14" />
    </svg>
  )
}

export function LockIcon({ label }: { label?: string }) {
  return (
    <svg className="icn icn-lock" viewBox="0 0 24 24" role={label ? 'img' : undefined} aria-label={label} aria-hidden={label ? undefined : true}>
      <rect x="5" y="11" width="14" height="10" rx="1" />
      <path d="M8 11V7a4 4 0 0 1 8 0v4" />
    </svg>
  )
}
