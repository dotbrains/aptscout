'use client';

import { Download, Terminal, Globe } from 'lucide-react';

export function HowItWorksSection() {
  const steps = [
    {
      icon: <Download className="w-8 h-8" />,
      step: '1',
      title: 'Install aptscout',
      description: 'Install via go install, Homebrew, or download a prebuilt binary from GitHub Releases. Single binary, zero runtime dependencies.',
    },
    {
      icon: <Terminal className="w-8 h-8" />,
      step: '2',
      title: 'Scrape Apartments',
      description: 'Run aptscout scrape to fetch all floor plans and available units. Data is stored in a local SQLite database with full price history.',
    },
    {
      icon: <Globe className="w-8 h-8" />,
      step: '3',
      title: 'Browse & Filter',
      description: 'Use aptscout list to filter from the CLI, or aptscout serve to browse in a local web UI with filters, sorting, and price charts.',
    },
  ];

  return (
    <section id="how-it-works" className="py-12 sm:py-16 lg:py-20 bg-dark-clay/50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-10 sm:mb-16">
          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-cream mb-3 sm:mb-4">
            How It Works
          </h2>
          <p className="text-cream/70 text-base sm:text-lg lg:text-xl max-w-3xl mx-auto">
            Three steps from install to browsing apartment listings
          </p>
        </div>
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6 sm:gap-8">
          {steps.map((step, index) => (
            <div key={index} className="relative sm:col-span-2 lg:col-span-1 last:sm:col-start-auto last:lg:col-start-auto">
              <div className="bg-dark-mesa border border-arizona-copper/30 rounded-xl p-6 sm:p-8 text-center hover:border-arizona-red/40 transition-all h-full">
                <div className="w-14 h-14 sm:w-16 sm:h-16 bg-gradient-to-br from-arizona-red to-arizona-copper rounded-full flex items-center justify-center text-white text-xl sm:text-2xl font-bold mx-auto mb-3 sm:mb-4">
                  {step.step}
                </div>
                <div className="w-10 h-10 sm:w-12 sm:h-12 mx-auto mb-3 sm:mb-4 text-arizona-red flex items-center justify-center">
                  {step.icon}
                </div>
                <h3 className="text-lg sm:text-xl font-semibold text-cream mb-2 sm:mb-3">{step.title}</h3>
                <p className="text-cream/60 text-sm sm:text-base leading-relaxed">{step.description}</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
