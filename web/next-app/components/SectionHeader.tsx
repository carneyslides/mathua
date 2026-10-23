interface SectionHeaderProps {
  label?: string
  title: string
  center?: boolean
  className?: string
}

export default function SectionHeader({ label, title, center = false, className = '' }: SectionHeaderProps) {
  return (
    <div className={`${center ? 'text-center' : ''} ${className}`}>
      {label && (
        <span className="font-mono text-[11px] uppercase tracking-[0.15em] text-mathua-muted block mb-2">
          {label}
        </span>
      )}
      <h2
        className="font-serif font-normal text-[1.9rem] max-sm:text-[1.5rem] leading-tight text-mathua-primary"
        style={{
          fontFamily: "var(--font-space-grotesk), 'Space Grotesk', serif",
          fontWeight: 400,
          color: 'var(--text-primary)',
          // A horizontal rule under the title, not a vertical bar beside it. The
          // bar was a second accent competing with the mono eyebrow above and it
          // read as a quote mark next to a page heading; the owner's own rule is
          // that this site does not use the left bar.
          paddingBottom: '0.75rem',
          borderBottom: '1px solid var(--border)',
          marginBottom: '1.5rem',
        }}
      >
        {title}
      </h2>
    </div>
  )
}
