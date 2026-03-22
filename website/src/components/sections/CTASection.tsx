'use client';

import { Github, BookOpen, MessageCircle } from 'lucide-react';

export function CTASection() {
  return (
    <section className="py-12 sm:py-16 lg:py-20 bg-gradient-to-br from-arizona-red/10 via-dark-mesa to-dark-mesa">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 text-center">
        <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-cream mb-4 sm:mb-6">
          Ready to Get Started?
        </h2>
        <p className="text-base sm:text-lg lg:text-xl text-cream/70 mb-8 sm:mb-12 max-w-3xl mx-auto">
          Install aptscout and start tracking apartment availability in under a minute
        </p>
        <div className="grid sm:grid-cols-3 gap-4 sm:gap-6">
          <a
            href="https://github.com/dotbrains/aptscout"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-dark-clay/50 border border-arizona-red/30 hover:border-arizona-red rounded-xl p-6 sm:p-8 transition-all group hover:shadow-lg hover:shadow-arizona-red/20"
          >
            <div className="w-12 h-12 sm:w-14 sm:h-14 bg-gradient-to-br from-arizona-red to-arizona-copper rounded-lg flex items-center justify-center mx-auto mb-3 sm:mb-4 group-hover:scale-110 transition-transform">
              <Github className="w-6 h-6 sm:w-7 sm:h-7 text-white" />
            </div>
            <h3 className="text-lg sm:text-xl font-semibold text-cream mb-2">View on GitHub</h3>
            <p className="text-cream/60 text-xs sm:text-sm">Star the repo, fork it, and contribute</p>
          </a>
          <a
            href="https://github.com/dotbrains/aptscout#readme"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-dark-clay/50 border border-arizona-copper/30 hover:border-arizona-copper rounded-xl p-6 sm:p-8 transition-all group hover:shadow-lg hover:shadow-arizona-copper/20"
          >
            <div className="w-12 h-12 sm:w-14 sm:h-14 bg-gradient-to-br from-arizona-copper to-arizona-gold rounded-lg flex items-center justify-center mx-auto mb-3 sm:mb-4 group-hover:scale-110 transition-transform">
              <BookOpen className="w-6 h-6 sm:w-7 sm:h-7 text-white" />
            </div>
            <h3 className="text-lg sm:text-xl font-semibold text-cream mb-2">Read the Docs</h3>
            <p className="text-cream/60 text-xs sm:text-sm">README, SPEC, and database schema</p>
          </a>
          <a
            href="https://github.com/dotbrains/aptscout/discussions"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-dark-clay/50 border border-arizona-gold/30 hover:border-arizona-gold rounded-xl p-6 sm:p-8 transition-all group hover:shadow-lg hover:shadow-arizona-gold/20"
          >
            <div className="w-12 h-12 sm:w-14 sm:h-14 bg-gradient-to-br from-arizona-gold to-arizona-red rounded-lg flex items-center justify-center mx-auto mb-3 sm:mb-4 group-hover:scale-110 transition-transform">
              <MessageCircle className="w-6 h-6 sm:w-7 sm:h-7 text-white" />
            </div>
            <h3 className="text-lg sm:text-xl font-semibold text-cream mb-2">Join Discussion</h3>
            <p className="text-cream/60 text-xs sm:text-sm">Ask questions and share ideas</p>
          </a>
        </div>
      </div>
    </section>
  );
}
