# OpenDI Model Hub Frontend

This frontend uses React with Vite.

## Scripts

- `npm start` - run the dev server on port 3000
- `npm run build` - create a production bundle in `dist/`
- `npm test` - run frontend tests with Vitest
- `npm run preview` - preview the production bundle locally

## API Configuration

The API base URL is provided by `VITE_API_URL`.

- Development compose sets `VITE_API_URL` from `API_URL`.
- Production build args set `VITE_API_URL` at image build time.
- If not provided, the app defaults to relative `/api` (reverse proxy friendly).
