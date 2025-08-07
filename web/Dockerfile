# Use a specific, stable Node.js version on Alpine for a small and secure base image.
FROM node:20-alpine

# Set the working directory inside the container.
WORKDIR /app

# Copy package manifests and install dependencies first to leverage Docker's build cache.
COPY package.json package-lock.json ./

# Install dependencies. Using 'npm ci' is recommended for CI/CD as it provides faster, more reliable builds.
RUN npm ci

# Copy the rest of the application source code into the container.
# This is done after npm install to avoid invalidating the cache on code changes.
COPY . .

# Expose the port the Vite dev server will run on.
EXPOSE 5173

# The command to start the Vite development server.
# Using 'npx' ensures the local vite binary is used.
CMD ["npx", "vite"] 