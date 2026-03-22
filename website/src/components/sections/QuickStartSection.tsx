'use client';

import React, { useState } from 'react';
import { CodeBlock } from '@/components/CodeBlock';

export function QuickStartSection() {
  const [installMethod, setInstallMethod] = useState<'go' | 'brew' | 'release'>('go');

  const goExample = `go install github.com/dotbrains/aptscout@latest`;

  const brewExample = `brew tap dotbrains/tap
brew install --cask aptscout`;

  const releaseExample = `# macOS Apple Silicon
gh release download --repo dotbrains/aptscout \\
  --pattern 'aptscout_darwin_arm64.tar.gz' --dir /tmp
tar -xzf /tmp/aptscout_darwin_arm64.tar.gz -C /usr/local/bin`;

  const installExamples = { go: goExample, brew: brewExample, release: releaseExample };

  return (
    <section id="quick-start" className="py-12 sm:py-16 lg:py-20 bg-dark-mesa overflow-hidden">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-cream mb-3 sm:mb-4">
            Quick Start
          </h2>
          <p className="text-canyon-slate text-base sm:text-lg lg:text-xl max-w-3xl mx-auto">
            Install aptscout and browse apartments in under a minute
          </p>
        </div>
        <div className="grid lg:grid-cols-2 gap-8 lg:gap-12 items-start">
          <div className="bg-dark-clay/50 rounded-xl p-6 sm:p-8 border border-arizona-copper/20 min-w-0">
            <h3 className="text-xl sm:text-2xl font-bold text-cream mb-4 sm:mb-6">1. Install</h3>
            <div className="flex gap-2 sm:gap-3 mb-6">
              {[
                { key: 'go' as const, label: 'Go' },
                { key: 'brew' as const, label: 'Homebrew' },
                { key: 'release' as const, label: 'Release' },
              ].map((method) => (
                <button
                  key={method.key}
                  onClick={() => setInstallMethod(method.key)}
                  className={`flex-1 px-3 sm:px-4 py-2.5 rounded-lg text-sm font-semibold transition-all ${
                    installMethod === method.key
                      ? 'bg-gradient-to-r from-arizona-red to-arizona-copper text-white shadow-lg shadow-arizona-red/30'
                      : 'bg-dark-mesa text-canyon-slate hover:text-cream hover:border-arizona-copper/50 border border-arizona-copper/30'
                  }`}
                >
                  {method.label}
                </button>
              ))}
            </div>
            <CodeBlock
              code={installExamples[installMethod]}
              language="bash"
            />
          </div>
          <div className="bg-dark-clay/50 rounded-xl p-6 sm:p-8 border border-arizona-red/20 min-w-0">
            <h3 className="text-xl sm:text-2xl font-bold text-cream mb-4 sm:mb-6">2. Use</h3>
            <CodeBlock
              code={`# Scrape all properties
aptscout scrape

# Scrape just one property
aptscout scrape --property hideaway

# List 2-bed units under $2,500
aptscout list --beds 2 --max-price 2500

# Filter by property
aptscout list --property desert-club

# Browse in the web UI
aptscout serve --open

# View summary stats
aptscout stats`}
              language="bash"
            />
            <div className="mt-6 bg-arizona-red/10 border border-arizona-red/30 rounded-lg p-4 sm:p-5">
              <p className="text-cream text-sm leading-relaxed">
                <span className="text-arizona-red font-semibold">Tip:</span> Schedule <code className="bg-dark-mesa/80 px-2 py-1 rounded text-arizona-gold font-mono text-xs">aptscout scrape</code> via cron to build a daily price history. Run <code className="bg-dark-mesa/80 px-2 py-1 rounded text-arizona-gold font-mono text-xs">aptscout serve</code> anytime to review trends.
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
