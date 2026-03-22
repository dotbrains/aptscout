import type { Metadata } from 'next';
import '@/styles/globals.css';

export const metadata: Metadata = {
  title: 'aptscout — Apartment Availability Tracker',
  description: 'Scrape apartment availability across multiple properties, store results in SQLite, and browse with a filterable web UI. Track prices, catch deals, never miss a unit.',
  openGraph: {
    title: 'aptscout — Apartment Availability Tracker',
    description: 'Scrape apartment availability across multiple properties, store results in SQLite, and browse with a filterable web UI. Track prices, catch deals, never miss a unit.',
    url: 'https://aptscout.dotbrains.io',
    siteName: 'aptscout',
    locale: 'en_US',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'aptscout — Apartment Availability Tracker',
    description: 'Scrape apartment availability across multiple properties, store results in SQLite, and browse with a filterable web UI. Track prices, catch deals, never miss a unit.',
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
