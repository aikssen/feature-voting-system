import { useState } from 'react'
import { Header } from './components/Header'
import { Hero } from './components/Hero'
import { Toolbar } from './components/Toolbar'
import { FeatureCard } from './components/FeatureCard'
import { Pager } from './components/Pager'
import { EmptyState, ErrorState, FeatureSkeleton } from './components/States'
import { AuthModal } from './auth/AuthModal'
import { SubmitFeatureForm } from './features/SubmitFeatureForm'
import { useFeatures } from './features/useFeatures'
import { useAuth } from './auth/useAuth'

export default function App() {
  const features = useFeatures()
  const { isAuthenticated, openAuth } = useAuth()
  const [submitOpen, setSubmitOpen] = useState(false)

  function handleSubmitClick() {
    if (!isAuthenticated) {
      openAuth('signup')
      return
    }
    setSubmitOpen(true)
    requestAnimationFrame(() =>
      document.getElementById('board')?.scrollIntoView({ behavior: 'smooth', block: 'start' }),
    )
  }

  function handleCreated() {
    features.setSort('newest')
    features.refresh()
  }

  return (
    <>
      <a
        href="#board"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[70] focus:rounded-lg focus:bg-accent focus:px-4 focus:py-2 focus:font-semibold focus:text-bg"
      >
        Skip to feature requests
      </a>
      <Header />
      <main>
        <Hero onSubmit={handleSubmitClick} />

        <section
          id="board"
          aria-labelledby="board-heading"
          className="mx-auto flex max-w-3xl scroll-mt-20 flex-col gap-5 px-4 pb-24 sm:px-6"
        >
          <h2 id="board-heading" className="sr-only">
            Feature requests
          </h2>

          <SubmitFeatureForm open={submitOpen} onClose={() => setSubmitOpen(false)} onCreated={handleCreated} />

          <Toolbar search={features.search} sort={features.sort} onSearch={features.setSearch} onSort={features.setSort} />

          <Results features={features} />
        </section>
      </main>

      <AuthModal />
    </>
  )
}

function Results({ features }: { features: ReturnType<typeof useFeatures> }) {
  const { data, loading, error, search, vote, setSearch, refresh, page, setPage } = features

  if (loading && !data) return <FeatureSkeleton />
  if (error && !data) return <ErrorState message={error.message} onRetry={refresh} />
  if (!data) return null

  if (data.items.length === 0) {
    return <EmptyState searching={search.trim().length > 0} onClear={() => setSearch('')} />
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="font-mono text-xs text-faint">
        {data.total} {data.total === 1 ? 'request' : 'requests'}
      </p>

      <div className={`flex flex-col gap-3 transition-opacity ${loading ? 'opacity-60' : 'opacity-100'}`}>
        {data.items.map((feature, i) => (
          <FeatureCard
            key={feature.id}
            feature={feature}
            onVote={vote}
            style={{ animationDelay: `${Math.min(i, 8) * 35}ms` }}
          />
        ))}
      </div>

      <Pager page={page} totalPages={data.total_pages} onPage={setPage} />
    </div>
  )
}
