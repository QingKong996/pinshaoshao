const COLUMNS = 12

export function GridOverlay() {
  return (
    <div
      aria-hidden
      data-slot="grid-overlay"
      className="pointer-events-none fixed inset-0 z-[100] mx-auto w-full max-w-6xl px-4 sm:px-6"
    >
      <div className="grid h-full grid-cols-12 gap-4">
        {Array.from({ length: COLUMNS }, (_, index) => (
          <div
            key={index}
            className="border-x border-dashed border-rose-500/40 bg-rose-500/5"
          />
        ))}
      </div>
    </div>
  )
}
