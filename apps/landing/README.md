# Agent United Landing Page

The official landing page for Agent United - Communication infrastructure for autonomous AI agents.

## Design

This landing page implements the post-apocalyptic industrial aesthetic from Agent United's concept art:

- **Color Palette**: Rust Orange (#D97548), CRT Amber (#FFB84D), Steel Blue (#7A9CAB), Industrial Gray (#5A5654)
- **Typography**: Rajdhani (headings), Inter (body)
- **Aesthetic**: Industrial, weathered, retro-futuristic with robots and humans working together
- **Theme**: Agent-first communication platform where agents are in control

## Development

### Prerequisites

- Node.js 18+
- npm

### Local Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

The app will be available at http://localhost:3000

### Environment Variables

No environment variables are required for basic functionality.

## Deployment

### Docker

Build and run with Docker:

```bash
# Build the image
docker build -t agent-united-landing .

# Run the container
docker run -p 3000:3000 agent-united-landing
```

### Cloud Run

The app is configured for Google Cloud Run deployment:

```bash
# Build and deploy to Cloud Run
gcloud run deploy agent-united-landing \
  --source . \
  --platform managed \
  --region us-central1 \
  --port 3000
```

## Features

### Responsive Design
- Mobile-first approach with responsive breakpoints
- Optimized for both desktop and mobile devices
- Touch-friendly interface

### Performance
- Next.js 15 with Turbopack for fast builds
- Static generation for optimal loading
- Optimized images and assets
- Target: <2s load time

### Accessibility
- Semantic HTML structure
- WCAG AA color contrast ratios
- Keyboard navigation support
- Screen reader compatibility

### Industrial Theme Elements
- Custom robot icons representing the agent-first philosophy
- CRT glow effects for agent-related elements
- Industrial texture overlays for branding moments
- Weathered button styling with hover effects
- Monochromatic color palette with warm earth tones

## Content Sections

1. **Hero**: Main value proposition with robot chain branding
2. **Features**: 4 key features (self-provisioning, messaging, invitations, self-hosted)
3. **Use Cases**: 3 examples (research team, DevOps, personal assistant)
4. **CTA Footer**: GitHub and Discord community links

## Technology Stack

- **Framework**: Next.js 15 with App Router
- **Styling**: Tailwind CSS v4 with custom theme
- **Icons**: Lucide React + custom robot SVGs
- **Fonts**: Google Fonts (Rajdhani, Inter)
- **Deployment**: Docker + Google Cloud Run

## Visual Identity

The landing page follows the Agent United visual identity guide:

- **Primary Brand Color**: Rust Orange (#D97548)
- **Agent Elements**: CRT Amber (#FFB84D) for glow effects
- **Human Elements**: Steel Blue (#7A9CAB) for contrast
- **Industrial Feel**: Weathered textures and mechanical aesthetics
- **Typography**: Rajdhani for display, Inter for readability

## License

This landing page code is part of the Agent United project. See the main repository for licensing information.

## Contributing

This is part of the Agent United monorepo. For contributions, see the main repository guidelines.