// Wire types — mirror the frozen contract in DECISIONS.md (Part B).

export interface UserView {
  id: string
  name: string
  email: string
  created_at: string
}

export interface AuthResponse {
  token: string
  user: UserView
}

export interface AuthorView {
  id: string
  name: string
}

export interface FeatureView {
  id: string
  title: string
  description: string
  author: AuthorView
  created_at: string
  total_votes: number
  trending_score: number
  has_voted: boolean
  is_author: boolean
  rank: number
}

export interface Page<T> {
  items: T[]
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
}

export interface VoteResponse {
  feature_id: string
  total_votes: number
  has_voted: boolean
}

export type Sort = 'trending' | 'most_voted' | 'newest'

export interface FieldError {
  field: string
  issue: string
}

export interface ApiErrorEnvelope {
  error: {
    code: string
    message: string
    details?: FieldError[]
  }
}

export interface ListParams {
  search?: string
  sort?: Sort
  page?: number
  limit?: number
}
