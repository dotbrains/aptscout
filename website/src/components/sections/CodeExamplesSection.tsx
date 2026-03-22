'use client';

import React, { useState } from 'react';
import { CodeBlock } from '@/components/CodeBlock';

export function CodeExamplesSection() {
  const [activeTab, setActiveTab] = useState<'scrape' | 'list' | 'history' | 'stats' | 'serve' | 'clean'>('scrape');

  const examples = {
    scrape: `# Scrape all properties
$ aptscout scrape

[1/2] Desert Club Apartments
  ⠋ Fetching floor plans and units...
  ✓ 14 plans, 24 units (24 new)

[2/2] Hideaway North Scottsdale
  ⠋ Fetching floor plans and units...
  ✓ 17 plans, 10 units (10 new)

✓ Scrape complete.
→ 34 units available (34 new)
→ Database: ~/.local/share/aptscout/aptscout.db

# Or scrape a single property
$ aptscout scrape --property desert-club`,
    list: `# List 2-bedroom apartments under $2,500 across all properties
$ aptscout list --beds 2 --max-price 2500

# Filter by property
$ aptscout list --property desert-club --beds 2 --max-price 2100

# Output as JSON for scripting
$ aptscout list --beds 2 --json

# Only renovated units available by April
$ aptscout list --renovated --available-by 2026-04-30 --sort price

# List registered properties
$ aptscout properties
  desert-club          Desert Club Apartments
  hideaway             Hideaway North Scottsdale`,
    history: `# Show price history for a specific unit
$ aptscout history 2146
UNIT #2146 — B2R (2 bed / 2 bath, 1,142 sq ft)

DATE                 PRICE     CHANGE
2026-03-15 10:00     $2,195    —
2026-03-18 10:00     $2,095    -$100
2026-03-22 18:00     $2,095    (no change)`,
    stats: `# Show summary statistics
$ aptscout stats
Desert Club Apartments — 6901 E Chauncey Lane, Phoenix, AZ 85054

Floor Plans:    14
Available Now:  12 units

By Bedrooms:
  1 bed:  4 units ($1,635 – $1,910)
  2 bed:  6 units ($2,010 – $2,215)
  3 bed:  2 units ($2,285 – $2,535)

Last Scrape:    2026-03-22 18:00:00 (2 minutes ago)
Total Scrapes:  15`,
    serve: `# Browse apartments in a local web UI
$ aptscout serve
→ Serving at http://localhost:8700
→ Database: ~/.local/share/aptscout/aptscout.db
→ Press Ctrl+C to stop

# Custom port + auto-open browser
$ aptscout serve --port 9000 --open
→ Serving at http://localhost:9000
→ Opening browser...`,
    clean: `# Remove apartments not seen in 14 days
$ aptscout clean --days 14
→ Removing 5 apartments not seen in 14 days...
✓ Cleaned 5 stale records.

# Preview what would be removed
$ aptscout clean --days 7 --dry-run
→ Would remove 8 apartments not seen in 7 days
→ #1055 (B1R), #2088 (A2P), #3112 (B2R), ...`,
  };

  const tabs = [
    { key: 'scrape' as const, label: 'Scrape', language: 'bash' },
    { key: 'list' as const, label: 'List & Filter', language: 'bash' },
    { key: 'history' as const, label: 'Price History', language: 'bash' },
    { key: 'stats' as const, label: 'Stats', language: 'bash' },
    { key: 'serve' as const, label: 'Web UI', language: 'bash' },
    { key: 'clean' as const, label: 'Clean', language: 'bash' },
  ];

  return (
    <section id="code-examples" className="py-12 sm:py-16 lg:py-20 bg-dark-clay/50">
      <div className="max-w-6xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-cream mb-3 sm:mb-4">
            Code Examples
          </h2>
          <p className="text-cream/70 text-base sm:text-lg lg:text-xl max-w-3xl mx-auto">
            See aptscout in action — scraping, filtering, price tracking, and more
          </p>
        </div>
        <div className="bg-dark-mesa border border-arizona-copper/30 rounded-xl overflow-hidden">
          <div className="flex border-b border-arizona-copper/30 overflow-x-auto">
            {tabs.map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={`flex-1 px-3 sm:px-6 py-3 sm:py-4 text-xs sm:text-sm font-semibold transition-colors whitespace-nowrap ${
                  activeTab === tab.key
                    ? 'bg-dark-clay/50 text-arizona-red border-b-2 border-arizona-red'
                    : 'text-cream/70 hover:text-cream hover:bg-dark-clay/30'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
          <div className="p-4 sm:p-6 overflow-x-auto">
            <CodeBlock
              code={examples[activeTab]}
              language={tabs.find((t) => t.key === activeTab)?.language}
            />
          </div>
        </div>
      </div>
    </section>
  );
}
