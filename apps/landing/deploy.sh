#!/bin/bash

# Agent United Landing Page Deployment Script
# Deploys to Google Cloud Run

set -e

echo "🚀 Deploying Agent United Landing Page to Cloud Run..."

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI is not installed. Please install it first."
    exit 1
fi

# Build and deploy
echo "📦 Building and deploying..."
gcloud run deploy agent-united-landing \
  --source . \
  --platform managed \
  --region us-central1 \
  --port 3000 \
  --allow-unauthenticated \
  --min-instances 0 \
  --max-instances 10 \
  --cpu 1 \
  --memory 512Mi \
  --timeout 300s

echo "✅ Deployment complete!"
echo "🌐 Your landing page should be available at the Cloud Run URL"