'use client';

export function StatsSection() {
  return (
    <section className="py-12 sm:py-16 bg-dark-clay/50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6">
        <div className="text-center mb-8 sm:mb-12">
          <h2 className="text-2xl sm:text-3xl lg:text-4xl font-bold text-cream mb-3 sm:mb-4">
            Stop clicking through dozens of floor plan pages
          </h2>
          <p className="text-cream/70 text-base sm:text-lg lg:text-xl">
            One command scrapes everything — filter, sort, and track changes locally
          </p>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 sm:gap-6 md:gap-8 text-center">
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-gradient mb-1 sm:mb-2">1–3 Bed</div>
            <div className="text-cream/60 text-sm sm:text-base">Configurations</div>
          </div>
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-gradient mb-1 sm:mb-2">Price</div>
            <div className="text-cream/60 text-sm sm:text-base">History Tracking</div>
          </div>
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-gradient mb-1 sm:mb-2">Web UI</div>
            <div className="text-cream/60 text-sm sm:text-base">Embedded SPA</div>
          </div>
          <div>
            <div className="text-2xl sm:text-3xl font-bold text-gradient mb-1 sm:mb-2">Zero</div>
            <div className="text-cream/60 text-sm sm:text-base">Dependencies</div>
          </div>
        </div>
      </div>
    </section>
  );
}
