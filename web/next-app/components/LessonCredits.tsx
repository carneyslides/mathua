import { concepts } from '../lib/conceptData'

/**
 * The attribution a lesson's licence requires on every page view of its content.
 *
 * ORCCA's Licensing section asks a redistributor to include, "in every digital format page view
 * (including but not limited to EPUB, PDF, and HTML) and on every physical printed page", the
 * sentence below. It is a licence term and not a courtesy: CC BY 4.0 section 3(a)(1) defers the
 * manner of attribution to the licensor, and this is the manner ORCCA specifies. `THIRD_PARTY_NOTICES.md`
 * quotes the clause whole.
 *
 * The sentence appears under the article in `/study` and under the step panel in `/learn`, because
 * both show ORCCA text — the lesson body in the first, the worked example lifted out of it in the
 * second. Nothing renders for a lesson whose publisher states its terms in prose instead: OpenStax
 * and Algebrica already carry their credit as the lesson body's first line, and neither asks for a
 * fixed sentence per view.
 *
 * The token comes from `data/lessons/mapping.json` through `scripts/merge_concepts`, which is why a
 * concept with no lesson gets none: a page cannot credit text it never shows.
 */
export default function LessonCredits({
  conceptId,
  className = '',
}: {
  conceptId: string
  className?: string
}) {
  const credit = concepts.find(c => c.id === conceptId)?.lesson_credit
  if (credit !== 'orcca') return null
  return (
    <p className={`mt-3 font-mono text-[11px] leading-relaxed text-mathua-muted ${className}`}>
      Original source material, products with readable and accessible math content and other
      information are freely available at{' '}
      <a href="https://pcc.edu/orcca" className="text-mathua-blue hover:underline">
        pcc.edu/orcca
      </a>
      .
    </p>
  )
}
