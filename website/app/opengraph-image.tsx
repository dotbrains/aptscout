import { ImageResponse } from 'next/og';

export const runtime = 'edge';
export const alt = 'aptscout — Apartment Availability Tracker';
export const size = { width: 1200, height: 630 };
export const contentType = 'image/png';

export default async function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'center',
          padding: '60px 80px',
          background: 'linear-gradient(135deg, #1a1714 0%, #120f0d 100%)',
          fontFamily: '-apple-system, BlinkMacSystemFont, sans-serif',
        }}
      >
        {/* Top accent bar */}
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            height: 4,
            background: 'linear-gradient(90deg, rgba(200,75,49,0.2), rgba(218,165,32,0.3), rgba(200,75,49,0.2))',
          }}
        />

        <div style={{ display: 'flex', alignItems: 'center', gap: 24, marginBottom: 24 }}>
          {/* Cactus logo */}
          <div
            style={{
              width: 80,
              height: 80,
              borderRadius: 16,
              background: 'linear-gradient(135deg, #C84B31, #B87333)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 40,
            }}
          >
            🌵
          </div>
          <div
            style={{
              fontSize: 64,
              fontWeight: 800,
              color: '#F5E6D3',
              letterSpacing: -1,
            }}
          >
            aptscout
          </div>
        </div>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            marginBottom: 40,
          }}
        >
          <div style={{ fontSize: 28, color: '#8B7D6B' }}>
            Apartment availability across multiple properties,
          </div>
          <div style={{ fontSize: 28, color: '#8B7D6B' }}>
            one command.
          </div>
        </div>

        {/* Badges */}
        <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
          {[
            { label: 'Multi-Property', color: '#C84B31' },
            { label: 'Price Tracking', color: '#B87333' },
            { label: 'SQLite Storage', color: '#DAA520' },
            { label: 'Web UI', color: '#8B7D6B' },
          ].map((badge) => (
            <div
              key={badge.label}
              style={{
                padding: '8px 20px',
                borderRadius: 20,
                background: '#2d2520',
                border: `1.5px solid ${badge.color}`,
                color: badge.color,
                fontSize: 16,
                fontWeight: 600,
              }}
            >
              {badge.label}
            </div>
          ))}
        </div>

        {/* Bottom accent bar */}
        <div
          style={{
            position: 'absolute',
            bottom: 0,
            left: 0,
            right: 0,
            height: 4,
            background: 'linear-gradient(90deg, rgba(200,75,49,0.2), rgba(218,165,32,0.3), rgba(200,75,49,0.2))',
          }}
        />
      </div>
    ),
    { ...size }
  );
}
