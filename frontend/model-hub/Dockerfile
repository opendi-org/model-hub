
# Base image 
FROM node:22

# Set working directory
WORKDIR /webapp

ENV REACT_APP_API_URL http://localhost:8080

# Install dependencies
COPY ./package*.json ./
RUN npm install

# Bundle app source
COPY . .

# Expose port 3000
EXPOSE 3000

# Start the application
CMD ["npm", "start"]