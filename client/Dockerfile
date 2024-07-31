# Use the official Node.js image
FROM node:current-slim

# Set the working directory
WORKDIR /app

# Copy the package.json and package-lock.json files to the working directory
COPY package*.json ./

RUN rm -rf node_modules
RUN rm -rf package-lock.json

# Install dependencies
RUN npm install --legacy-peer-deps

# Copy the rest of the application files to the working directory
COPY . .

# Expose the port the app runs on
EXPOSE 3456

# Accept build argument
ARG BACKEND_URL

# Set environment variable from the build argument
ENV BACKEND_URL=${BACKEND_URL}

# Command to start the application in development mode
CMD ["npm", "run", "dev"]
