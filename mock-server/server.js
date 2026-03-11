import express from 'express'
import cors from 'cors'

const app = express()
app.use(cors())
app.use(express.json())

// Delay middleware for loading state testing (set SLOW=1 env var)
const SLOW_MODE = process.env.SLOW === '1'
const DELAY_MS = 1500

if (SLOW_MODE) {
  app.use(async (req, res, next) => {
    if (req.path.startsWith('/api') || req.path.startsWith('/rss')) {
      await new Promise(resolve => setTimeout(resolve, DELAY_MS))
    }
    next()
  })
}

const tools = [
  { id: 1, slug: 'claude-code', name: 'Claude Code', source_type: 'github', source_repo: 'anthropics/claude-code', homepage: 'https://claude.ai/code', is_active: 1 },
  { id: 2, slug: 'codex', name: 'OpenAI Codex', source_type: 'github', source_repo: 'openai/codex', homepage: 'https://github.com/openai/codex', is_active: 1 },
  { id: 3, slug: 'gemini-cli', name: 'Gemini CLI', source_type: 'github', source_repo: 'google-gemini/gemini-cli', homepage: 'https://github.com/google-gemini/gemini-cli', is_active: 1 },
  { id: 4, slug: 'opencode', name: 'OpenCode', source_type: 'github', source_repo: 'opencode-ai/opencode', homepage: 'https://github.com/opencode-ai/opencode', is_active: 1 }
]

// Generate more entries for pagination testing
const generateEntries = () => {
  const baseEntries = [
    {
      id: 1, tool_id: 1, tool_slug: 'claude-code', tool_name: 'Claude Code',
      version: 'v1.0.0', title: 'Initial Release', url: 'https://github.com/anthropics/claude-code/releases/tag/v1.0.0',
      body_md: '## Claude Code v1.0.0\n\nFirst stable release of Claude Code!\n\n### Features\n- AI-powered code completion\n- Multi-file editing\n- Context-aware suggestions\n\n### Bug Fixes\n- Fixed memory leak in tokenizer\n- Improved startup time',
      excerpt: '## Claude Code v1.0.0\n\nFirst stable release of Claude Code!...',
      published_at: '2026-03-10T10:00:00Z', is_prerelease: 0
    },
    {
      id: 2, tool_id: 1, tool_slug: 'claude-code', tool_name: 'Claude Code',
      version: 'v0.9.5', title: 'Beta Release', url: 'https://github.com/anthropics/claude-code/releases/tag/v0.9.5',
      body_md: '## Claude Code v0.9.5\n\nBeta release with improved features.\n\n### New\n- Dark mode support\n- Better error handling',
      excerpt: '## Claude Code v0.9.5\n\nBeta release with improved features...',
      published_at: '2026-03-08T14:30:00Z', is_prerelease: 1
    },
    {
      id: 3, tool_id: 2, tool_slug: 'codex', tool_name: 'OpenAI Codex',
      version: 'v2.1.0', title: 'Codex Update', url: 'https://github.com/openai/codex/releases/tag/v2.1.0',
      body_md: '## Codex v2.1.0\n\nMajor update with new capabilities.\n\n### Highlights\n- 50% faster inference\n- Support for 20+ languages\n- Better documentation generation',
      excerpt: '## Codex v2.1.0\n\nMajor update with new capabilities...',
      published_at: '2026-03-09T09:00:00Z', is_prerelease: 0
    },
    {
      id: 4, tool_id: 3, tool_slug: 'gemini-cli', tool_name: 'Gemini CLI',
      version: 'v0.5.0', title: 'Gemini CLI Alpha', url: 'https://github.com/google-gemini/gemini-cli/releases/tag/v0.5.0',
      body_md: '## Gemini CLI v0.5.0\n\nAlpha release.\n\nInitial public alpha with basic functionality.',
      excerpt: '## Gemini CLI v0.5.0\n\nAlpha release...',
      published_at: '2026-03-07T16:00:00Z', is_prerelease: 1
    },
    {
      id: 5, tool_id: 4, tool_slug: 'opencode', tool_name: 'OpenCode',
      version: 'v1.2.3', title: 'Stable Release', url: 'https://github.com/opencode-ai/opencode/releases/tag/v1.2.3',
      body_md: '## OpenCode v1.2.3\n\nStable release with bug fixes.\n\n### Fixed\n- Performance improvements\n- Memory optimization',
      excerpt: '## OpenCode v1.2.3\n\nStable release with bug fixes...',
      published_at: '2026-03-06T11:00:00Z', is_prerelease: 0
    }
  ]

  // Add more entries for pagination (total 30 entries)
  const toolConfigs = [
    { tool_id: 1, slug: 'claude-code', name: 'Claude Code' },
    { tool_id: 2, slug: 'codex', name: 'OpenAI Codex' },
    { tool_id: 3, slug: 'gemini-cli', name: 'Gemini CLI' },
    { tool_id: 4, slug: 'opencode', name: 'OpenCode' }
  ]

  for (let i = 6; i <= 30; i++) {
    const toolIndex = i % 4
    const tool = toolConfigs[toolIndex]
    const patchVersion = Math.floor(i / 4)
    const minorVersion = i % 10

    baseEntries.push({
      id: i,
      tool_id: tool.tool_id,
      tool_slug: tool.slug,
      tool_name: tool.name,
      version: `v0.${minorVersion}.${patchVersion}`,
      title: `Release ${i}`,
      url: `https://github.com/example/${tool.slug}/releases/tag/v0.${minorVersion}.${patchVersion}`,
      body_md: `## ${tool.name} v0.${minorVersion}.${patchVersion}\n\nRelease number ${i}.\n\n### Changes\n- Bug fix ${i}\n- Improvement ${i}`,
      excerpt: `## ${tool.name} v0.${minorVersion}.${patchVersion}\n\nRelease number ${i}...`,
      published_at: new Date(2026, 2, 10 - i, 10, 0, 0).toISOString(),
      is_prerelease: i % 5 === 0 ? 1 : 0
    })
  }

  return baseEntries
}

const entries = generateEntries()

// Health check
app.get('/api/health', (req, res) => {
  res.json({ status: 'ok', database: 'connected' })
})

// Get all tools
app.get('/api/tools', (req, res) => {
  res.json(tools)
})

// Get single tool
app.get('/api/tools/:slug', (req, res) => {
  const tool = tools.find(t => t.slug === req.params.slug)
  if (!tool) {
    return res.status(404).json({ error: 'tool not found' })
  }
  res.json(tool)
})

// Get all entries
app.get('/api/entries', async (req, res) => {
  const { tool, cursor, slow } = req.query
  let filtered = [...entries]

  // Optional delay for loading state testing
  if (slow === '1') {
    await new Promise(resolve => setTimeout(resolve, 2000))
  }

  if (tool) {
    filtered = filtered.filter(e => e.tool_slug === tool)
  }

  if (cursor) {
    const cursorIndex = filtered.findIndex(e => e.id === parseInt(cursor))
    if (cursorIndex !== -1) {
      filtered = filtered.slice(cursorIndex + 1)
    }
  }

  const limit = 20
  const hasMore = filtered.length > limit
  const result = filtered.slice(0, limit)

  res.json({
    entries: result,
    hasMore,
    nextCursor: hasMore && result.length > 0 ? result[result.length - 1].id : null
  })
})

// Get single entry
app.get('/api/entries/:id', (req, res) => {
  const entry = entries.find(e => e.id === parseInt(req.params.id))
  if (!entry) {
    return res.status(404).json({ error: 'entry not found' })
  }
  res.json(entry)
})

// Get tool entries
app.get('/api/tools/:slug/entries', (req, res) => {
  const tool = tools.find(t => t.slug === req.params.slug)
  if (!tool) {
    return res.status(404).json({ error: 'tool not found' })
  }

  const { cursor } = req.query
  let filtered = entries.filter(e => e.tool_slug === req.params.slug)

  if (cursor) {
    const cursorIndex = filtered.findIndex(e => e.id === parseInt(cursor))
    if (cursorIndex !== -1) {
      filtered = filtered.slice(cursorIndex + 1)
    }
  }

  const limit = 20
  const hasMore = filtered.length > limit
  const result = filtered.slice(0, limit)

  res.json({
    tool,
    entries: result,
    hasMore,
    nextCursor: hasMore && result.length > 0 ? result[result.length - 1].id : null
  })
})

// Search
app.get('/api/search', (req, res) => {
  const { q } = req.query
  if (!q) {
    return res.status(400).json({ error: 'query parameter required' })
  }

  const results = entries.filter(e =>
    e.title.toLowerCase().includes(q.toLowerCase()) ||
    e.body_md.toLowerCase().includes(q.toLowerCase())
  )

  res.json({ query: q, entries: results })
})

// RSS - All
app.get('/rss/all', (req, res) => {
  const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Agent Tracker - All Releases</title>
    <link>http://localhost:8080</link>
    <description>All AI coding tool releases</description>
    ${entries.map(e => `
    <item>
      <title>${e.title}</title>
      <link>${e.url}</link>
      <description>${e.excerpt}</description>
      <pubDate>${new Date(e.published_at).toUTCString()}</pubDate>
      <guid>${e.url}</guid>
    </item>`).join('')}
  </channel>
</rss>`

  res.set('Content-Type', 'application/xml; charset=utf-8')
  res.send(rss)
})

// RSS - Tool specific
app.get('/rss/:slug', (req, res) => {
  const tool = tools.find(t => t.slug === req.params.slug)
  if (!tool) {
    return res.status(404).json({ error: 'tool not found' })
  }

  const toolEntries = entries.filter(e => e.tool_slug === req.params.slug)

  const rss = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Agent Tracker - ${tool.name} Releases</title>
    <link>http://localhost:8080</link>
    <description>${tool.name} releases</description>
    ${toolEntries.map(e => `
    <item>
      <title>${e.title}</title>
      <link>${e.url}</link>
      <description>${e.excerpt}</description>
      <pubDate>${new Date(e.published_at).toUTCString()}</pubDate>
      <guid>${e.url}</guid>
    </item>`).join('')}
  </channel>
</rss>`

  res.set('Content-Type', 'application/xml; charset=utf-8')
  res.send(rss)
})

const PORT = process.env.PORT || 3001
app.listen(PORT, () => {
  console.log(`Mock API server running on port ${PORT}`)
})