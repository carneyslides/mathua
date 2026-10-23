import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import LessonCredits from '../components/LessonCredits'

/**
 * ORCCA's licence asks for one fixed sentence on every page view of its content, and a page view is
 * whichever surface renders that content: the article in `/study` and the step panel in `/learn`.
 * The sentence is quoted verbatim in `THIRD_PARTY_NOTICES.md`, so it is asserted here rather than
 * left to the component's own wording.
 */
const SENTENCE =
  'Original source material, products with readable and accessible math content and other information are freely available at pcc.edu/orcca.'

describe('LessonCredits', () => {
  it('renders the sentence ORCCA requires, with its link', () => {
    const { container } = render(<LessonCredits conceptId="alg.linear.slope" />)
    expect(container.textContent?.replace(/\s+/g, ' ').trim()).toBe(SENTENCE)
    expect(screen.getByRole('link')).toHaveAttribute('href', 'https://pcc.edu/orcca')
  })

  it('renders nothing for a lesson whose terms are stated in prose', () => {
    const { container } = render(<LessonCredits conceptId="abstract.group.def" />)
    expect(container).toBeEmptyDOMElement()
  })

  it('renders nothing for a concept with no lesson to credit', () => {
    const { container } = render(<LessonCredits conceptId="no.such.concept" />)
    expect(container).toBeEmptyDOMElement()
  })
})
