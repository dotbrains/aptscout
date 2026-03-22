import type { Metadata } from 'next';
import '@/styles/globals.css';

export const metadata: Metadata = {
  title: 'aptscout — Apartment Availability Tracker CLI',
  description: 'Scrape apartment availability from Desert Club Apartments, store results in SQLite, and browse with a filterable web UI. Scrape once, browse anytime.',
  openGraph: {
    title: 'aptscout — Apartment Availability Tracker CLI',
    description: 'Scrape apartment availability from Desert Club Apartments, store results in SQLite, and browse with a filterable web UI. Scrape once, browse anytime.',
    url: 'https://aptscout.dotbrains.io',
    siteName: 'aptscout',
    images: [
      {
        url: '/og-image.svg',
        width: 1200,
        height: 630,
        alt: 'aptscout — Apartment Availability Tracker CLI',
      },
    ],
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'aptscout — Apartment Availability Tracker CLI',
    description: 'Scrape apartment availability from Desert Club Apartments, store results in SQLite, and browse with a filterable web UI. Scrape once, browse anytime.',
    images: ['/og-image.svg'],
  },
  icons: {
    icon: [
      {
        url: '/favicon.svg',
        type: 'image/svg+xml',
      },
    ],
    apple: [
      {
        url: '/favicon.svg',
        type: 'image/svg+xml',
      },
    ],
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      </head>
      <body>{children}</body>
    </html>
  );
}
