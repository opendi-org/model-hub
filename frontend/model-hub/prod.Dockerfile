# Stage 1: Build the React app
FROM node:22-alpine AS build

WORKDIR /webapp

# Copy package.json and package-lock.json
COPY ./package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the app’s source code
COPY . .

# Build the app
RUN npm run build

# Stage 2: Serve the app with a lightweight image
FROM node:22-alpine


WORKDIR /webapp

# Set environment variable (only needed for build time)
ENV REACT_APP_API_URL=http://129.213.115.50:8080

# Install serve globally
RUN npm install -g serve

# Copy the build files from the previous stage
COPY --from=build /webapp/build ./build

# Set the command to run the app
CMD ["serve", "-s", "build", "-l", "3000"]

# Expose the port
EXPOSE 3000