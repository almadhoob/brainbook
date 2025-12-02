# BrainBook Frontend

This package contains the Nuxt 4 dashboard that powers the BrainBook social network. It delivers the authenticated experience (home, posts, groups, users, messages, notifications, and settings) to end users.

## API configuration

The UI talks to the Go backend through the `NUXT_PUBLIC_API_BASE` runtime config. During local development the default of `http://localhost:8080` (configured in `nuxt.config.ts`) matches the backend server. To point the UI at a different host:

```bash
NUXT_PUBLIC_API_BASE="https://brainbook.vercel.app" pnpm dev
```

or update the `runtimeConfig.public.apiBase` value in `nuxt.config.ts`.

## Setup

Make sure to install the dependencies:

```bash
pnpm install
```

## Development

Start the development server on `http://localhost:3000`:

```bash
pnpm dev
```

## Production

Build the application for production:

```bash
pnpm build
```

Locally preview production build:

```bash
pnpm preview
```

Check out the [deployment documentation](https://nuxt.com/docs/getting-started/deployment) for more information.
