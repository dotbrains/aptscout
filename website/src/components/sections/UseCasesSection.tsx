'use client';

import { Search, TrendingDown, Globe, Bell, Filter, Clock, Repeat, Database } from 'lucide-react';

export function UseCasesSection() {
  const useCases = [
    {
      icon: <Search className="w-6 h-6" />,
      title: 'Apartment Shopping',
      description: 'See all available units across every floor plan in one view. Filter by beds, price, sqft — no more clicking through 14 separate pages.',
    },
    {
      icon: <TrendingDown className="w-6 h-6" />,
      title: 'Price Drop Hunting',
      description: 'Track price changes over time. Catch when a unit drops $100 overnight and act fast before someone else does.',
    },
    {
      icon: <Globe className="w-6 h-6" />,
      title: 'Visual Browsing',
      description: 'Use the built-in web UI to browse with filters, sort controls, and price history charts. Dark theme, keyboard shortcuts included.',
    },
    {
      icon: <Bell className="w-6 h-6" />,
      title: 'New Unit Alerts',
      description: 'Re-scrape daily and check the summary. Instantly see how many new units appeared and which ones disappeared.',
    },
    {
      icon: <Filter className="w-6 h-6" />,
      title: 'Budget Filtering',
      description: 'Set your max price and bedroom count once. aptscout list shows only what fits your budget — no distractions.',
    },
    {
      icon: <Clock className="w-6 h-6" />,
      title: 'Move-in Planning',
      description: 'Filter by availability date to find units available when you need them. Plan your move-in timeline with exact dates.',
    },
    {
      icon: <Repeat className="w-6 h-6" />,
      title: 'Cron Automation',
      description: 'Schedule aptscout scrape via cron to build a daily record. Review trends over weeks with the web UI or stats command.',
    },
    {
      icon: <Database className="w-6 h-6" />,
      title: 'Data Export',
      description: 'Use --json output to pipe data to jq, import into spreadsheets, or feed into custom scripts for advanced analysis.',
    },
  ];

  return (
    <section id="use-cases" className="py-12 sm:py-16 lg:py-20 bg-dark-mesa">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-cream mb-3 sm:mb-4">
            Use Cases
          </h2>
          <p className="text-cream/70 text-base sm:text-lg lg:text-xl max-w-3xl mx-auto">
            aptscout adapts to however you hunt for apartments
          </p>
        </div>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6 lg:gap-8">
          {useCases.map((useCase, index) => (
            <div
              key={index}
              className="bg-dark-clay/50 border border-arizona-copper/20 rounded-xl p-5 sm:p-6 hover:border-arizona-red/40 transition-all"
            >
              <div className="w-10 h-10 sm:w-12 sm:h-12 bg-gradient-to-br from-arizona-red to-arizona-copper rounded-lg flex items-center justify-center text-white mb-3 sm:mb-4">
                {useCase.icon}
              </div>
              <h3 className="text-lg sm:text-xl font-semibold text-cream mb-2">{useCase.title}</h3>
              <p className="text-cream/60 text-sm sm:text-base leading-relaxed">{useCase.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
