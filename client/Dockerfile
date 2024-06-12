# Use the official Node.js image
FROM node:latest

# Set the working directory
WORKDIR /app

# Copy the package.json and package-lock.json files to the working directory
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install --legacy-peer-deps

# Copy the rest of the application files to the working directory
COPY . .

# Expose the port the app runs on
EXPOSE 5173

# Set the command to start the application in development mode
CMD ["npm", "run", "dev"]
