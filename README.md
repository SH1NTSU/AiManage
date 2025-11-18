# 🤖 AiManage

> A comprehensive AI model training and management platform with an innovative dual-mode training system, community marketplace, and HuggingFace integration.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.24.3-blue.svg)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18.3.1-61DAFB.svg)](https://reactjs.org/)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB.svg)](https://www.python.org/)

---

## 📋 Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Database Setup](#database-setup)
  - [Running the Application](#running-the-application)
- [Training Modes](#training-modes)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Subscription Tiers](#subscription-tiers)
- [HuggingFace Integration](#huggingface-integration)
- [Contributing](#contributing)
- [License](#license)

---

## 🎯 Overview

**AiManage** is a full-stack AI model training and management platform that enables users to train, manage, publish, and share machine learning models. The platform features a unique **dual-mode training system** where users can either train models locally on their own hardware (free) or use cloud resources (paid subscriptions).

### Why AiManage?

- **🆓 Free Local Training**: Train models on your own hardware at no cost
- **☁️ Cloud Training**: Scale up with GPU-accelerated cloud training (T4, V100, A100)
- **🏪 Community Marketplace**: Publish, share, and monetize your trained models
- **🔄 HuggingFace Integration**: Seamlessly import/export models from HuggingFace Hub
- **📊 Real-time Monitoring**: Live training progress with WebSocket streaming
- **📈 Advanced Analytics**: Comprehensive training metrics and overfitting detection
- **💳 Flexible Pricing**: From free local training to enterprise cloud solutions

---

## ✨ Key Features

### 🎓 Dual-Mode Training System

#### Free Tier - Local Training
- Run a training agent on your own machine
- Use your own compute resources
- Full web interface for management and monitoring
- Real-time progress streaming via WebSocket
- No subscription required

#### Paid Tier - Server Training
- Training on platform's cloud infrastructure
- GPU access (T4, V100, A100 depending on tier)
- Subscription-based pricing ($9.99 - $99.99/month)
- Training credits system
- No local setup required

### 🏪 Community Marketplace

- **Publish & Share**: Make your models available to the community
- **Monetization**: Offer models for free or set your own price
- **Licensing Options**: Personal use, commercial, MIT, Apache 2.0
- **Social Features**: Likes, comments, ratings on models
- **Categories & Tags**: Organize and discover models easily
- **Featured Models**: Highlighting exceptional models
- **Download Tracking**: Monitor model popularity and usage

### 🔄 HuggingFace Integration

- Push trained models to HuggingFace Hub
- Import existing models from HuggingFace
- Search HuggingFace model repository
- Inference API integration
- Automatic model card generation

### 📊 Training Features

- **Real-time Metrics**: Loss, accuracy, validation metrics
- **Comprehensive Analytics**: Overfitting/underfitting detection
- **WebSocket Streaming**: Live training logs and progress
- **Training History**: Track all epochs and metrics
- **Model Analysis**: Detailed performance insights
- **Multiple Modes**: Local paths or server uploads

### 🔐 Authentication & Security

- Email/password authentication
- OAuth providers: Google, GitHub, Apple Sign In
- JWT-based session management with refresh tokens
- API keys for training agent authentication
- Secure password validation

### 💳 Subscription Management

- Four subscription tiers (Free, Basic, Pro, Enterprise)
- Stripe integration for payments
- Training credits system
- Automatic subscription updates via webhooks
- Mock mode for development/testing

---

## 🏗️ Architecture

AiManage consists of three main components:

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │      │                 │
│  React Frontend │◄────►│   Go Backend    │◄────►│  PostgreSQL DB  │
│   (Port 8080)   │      │   (Port 8000)   │      │   (Port 5432)   │
│                 │      │                 │      │                 │
└─────────────────┘      └────────┬────────┘      └─────────────────┘
                                  │
                                  │ WebSocket
                                  │
                         ┌────────▼────────┐
                         │                 │
                         │ Python Training │
                         │     Agent       │
                         │                 │
                         └─────────────────┘
```

### Component Breakdown

1. **Frontend (React + TypeScript)**
   - Modern UI with shadcn/ui components
   - Real-time updates via WebSocket
   - TanStack Query for data management
   - Stripe checkout integration

2. **Backend (Go)**
   - Chi router for HTTP routing
   - PostgreSQL with pgx driver
   - JWT authentication
   - WebSocket server for real-time updates
   - Stripe webhook handling

3. **Training Agent (Python)**
   - PyTorch-based training engine
   - WebSocket client for communication
   - HuggingFace Transformers integration
   - Real-time metric reporting

---

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.24.3
- **Framework**: Chi Router v5
- **Database**: PostgreSQL 16 (pgx/v5)
- **Authentication**: JWT with refresh tokens
- **Real-time**: WebSockets (gorilla/websocket)
- **Payments**: Stripe v81
- **AI Integration**: Claude AI API

### Frontend
- **Framework**: React 18.3.1 + TypeScript
- **Build Tool**: Vite 5.4.19
- **UI Library**: Radix UI + Tailwind CSS + shadcn/ui
- **State Management**: React Context API
- **Data Fetching**: TanStack Query v5
- **Routing**: React Router v6
- **Charts**: Recharts v2
- **Authentication**: JWT + Google OAuth
- **Payments**: Stripe React

### Training Agent
- **Language**: Python 3.8+
- **ML Framework**: PyTorch
- **Communication**: WebSockets (asyncio)
- **ML Integration**: HuggingFace Hub, Transformers

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.24.3 or higher
- **Node.js** 18+ and npm/yarn
- **Python** 3.8 or higher
- **PostgreSQL** 16
- **Docker** (optional, for database)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/AiManage.git
   cd AiManage
   ```

2. **Backend Setup**
   ```bash
   cd server

   # Install Go dependencies
   go mod download

   # Copy environment file
   cp .env.example .env

   # Edit .env with your configuration
   nano .env
   ```

3. **Frontend Setup**
   ```bash
   cd app

   # Install dependencies
   npm install

   # Copy environment file
   cp .env.example .env

   # Edit .env with your configuration
   nano .env
   ```

4. **Training Agent Setup**
   ```bash
   cd training-agent

   # Create virtual environment
   python3 -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate

   # Install dependencies
   pip install -r requirements.txt
   ```

### Database Setup

#### Option 1: Using Docker (Recommended)

```bash
cd server
docker-compose up -d
```

This will start PostgreSQL on port 5432 and pgAdmin on port 5050.

#### Option 2: Local PostgreSQL

1. Install PostgreSQL 16
2. Create a database:
   ```bash
   createdb aimanage
   ```

3. Run migrations:
   ```bash
   cd server
   make migrate-up
   ```

### Running the Application

1. **Start the Backend**
   ```bash
   cd server
   go run cmd/server/main.go
   ```
   Server will run on `http://localhost:8000`

2. **Start the Frontend**
   ```bash
   cd app
   npm run dev
   ```
   Frontend will run on `http://localhost:8080`

3. **Start the Training Agent** (for local training)
   ```bash
   cd training-agent
   source venv/bin/activate
   python train_agent.py --api-key YOUR_API_KEY
   ```

### Environment Variables

#### Backend (.env)
```env
DB_URI=postgresql://postgres:postgres@localhost:5432/aimanage
JWT_SECRET=your-secret-key
GEMINI_API_KEY=your-gemini-api-key
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
APPLE_CLIENT_ID=your-apple-client-id
APPLE_CLIENT_SECRET=your-apple-client-secret
STRIPE_SECRET_KEY=your-stripe-secret-key
STRIPE_WEBHOOK_SECRET=your-stripe-webhook-secret
```

#### Frontend (.env)
```env
VITE_GOOGLE_CLIENT_ID=your-google-client-id
VITE_GITHUB_CLIENT_ID=your-github-client-id
VITE_APPLE_CLIENT_ID=your-apple-client-id
VITE_STRIPE_PUBLISHABLE_KEY=your-stripe-publishable-key
```

For detailed setup instructions, see:
- [OAuth Setup Guide](./AUTH_SETUP.md)
- [Stripe Setup Guide](./STRIPE_SETUP.md)
- [HuggingFace Setup Guide](./HUGGINGFACE_SETUP.md)

---

## 🎯 Training Modes

### Local Training (Free)

Perfect for hobbyists, students, and developers who want to experiment without cost.

1. Generate an API key in the Settings page
2. Download and run the training agent on your machine
3. Upload your training script and dataset via the web interface
4. Start training and monitor progress in real-time

```bash
# Run the training agent
cd training-agent
python train_agent.py --api-key YOUR_API_KEY
```

**Benefits:**
- ✅ Completely free
- ✅ Use your own hardware
- ✅ No training limits
- ✅ Full control over resources

### Cloud Training (Paid)

Ideal for professionals and teams who need scalable, GPU-accelerated training.

1. Subscribe to a paid tier (Basic, Pro, or Enterprise)
2. Upload your training script and dataset
3. Start training on cloud infrastructure
4. Monitor progress and download results

**Benefits:**
- ✅ GPU acceleration (T4, V100, A100)
- ✅ No local setup required
- ✅ Scalable compute resources
- ✅ Priority support

---

## 📁 Project Structure

```
AiManage/
├── app/                          # React Frontend
│   ├── src/
│   │   ├── components/           # React components
│   │   │   ├── ui/              # shadcn/ui components (40+)
│   │   │   ├── Layout.tsx
│   │   │   ├── AppSidebar.tsx
│   │   │   ├── HuggingFaceIntegration.tsx
│   │   │   ├── StripeCheckout.tsx
│   │   │   └── ProtectedRoute.tsx
│   │   ├── pages/               # Page components
│   │   │   ├── Models.tsx
│   │   │   ├── Community.tsx
│   │   │   ├── ModelDetail.tsx
│   │   │   ├── Settings.tsx
│   │   │   ├── Statistics.tsx
│   │   │   ├── Pricing.tsx
│   │   │   └── Auth.tsx
│   │   ├── context/             # React contexts
│   │   │   ├── authContext.tsx
│   │   │   ├── modelContext.tsx
│   │   │   ├── trainingContext.tsx
│   │   │   └── subscriptionContext.tsx
│   │   └── lib/                 # Utilities
│   └── package.json
│
├── server/                       # Go Backend
│   ├── cmd/server/main.go       # Entry point
│   ├── internal/
│   │   ├── handlers/            # HTTP handlers
│   │   │   ├── auth.go
│   │   │   ├── oauth.go
│   │   │   ├── training.go
│   │   │   ├── insertModel.go
│   │   │   ├── publishHandler.go
│   │   │   ├── communityHandler.go
│   │   │   ├── subscription.go
│   │   │   ├── huggingface.go
│   │   │   └── user.go
│   │   ├── service/
│   │   │   ├── router.go
│   │   │   ├── wsServer.go
│   │   │   └── trainingWS.go
│   │   ├── repository/
│   │   │   ├── model.go
│   │   │   └── subscription.go
│   │   ├── middlewares/
│   │   │   ├── auth.go
│   │   │   └── cors.go
│   │   └── types/
│   │       └── main.go
│   ├── aiAgent/
│   │   ├── trainer.go
│   │   └── metrics.go
│   ├── helpers/
│   │   ├── jwt.go
│   │   ├── huggingface.go
│   │   └── zip.go
│   ├── migrations/              # Database migrations
│   ├── go.mod
│   ├── Makefile
│   └── docker-compose.yml
│
├── training-agent/              # Python Training Agent
│   ├── train_agent.py
│   ├── requirements.txt
│   └── README.md
│
├── training-cli/                # Standalone CLI tools
│   ├── train_local.py
│   ├── upload_model.py
│   └── README.md
│
└── demo_model/                  # Example training project
    ├── train.py
    ├── model.py
    ├── requirements.txt
    └── data/
```

---

## 📡 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/register` | Create new account |
| POST | `/v1/login` | User login |
| GET | `/v1/refresh` | Refresh JWT token |
| POST | `/v1/auth/google` | Google OAuth |
| POST | `/v1/auth/github` | GitHub OAuth |
| POST | `/v1/auth/apple` | Apple Sign In |

### Model Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/insert` | Upload new model |
| GET | `/v1/getModels` | List user's models |
| DELETE | `/v1/deleteModel` | Delete model |
| GET | `/v1/downloadModel` | Download trained model |

### Training Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/train/start` | Start training session |
| GET | `/v1/train/progress` | Get training progress |
| POST | `/v1/train/analyze` | Analyze training results |
| POST | `/v1/train/cleanup` | Cleanup old trainings |

### Community Marketplace

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/publish` | Publish model to marketplace |
| POST | `/v1/published-models/{id}/unpublish` | Unpublish model |
| GET | `/v1/published-models` | List all published models |
| GET | `/v1/my-published-models` | User's published models |
| GET | `/v1/published-models/{id}` | Get model details |
| POST | `/v1/published-models/{id}/download` | Download published model |
| POST | `/v1/published-models/{id}/like` | Like a model |
| DELETE | `/v1/published-models/{id}/like` | Unlike a model |
| GET | `/v1/published-models/{id}/comments` | Get comments |
| POST | `/v1/published-models/{id}/comments` | Add comment |
| DELETE | `/v1/comments/{commentId}` | Delete comment |

### Subscription Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/subscription` | Get subscription info |
| POST | `/v1/subscription/checkout` | Create Stripe checkout |
| POST | `/v1/subscription/mock-upgrade` | Mock upgrade (dev) |
| GET | `/v1/pricing` | Get pricing tiers |
| POST | `/v1/webhook/stripe` | Stripe webhook handler |

### HuggingFace Integration

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/huggingface/push` | Push to HuggingFace Hub |
| POST | `/v1/huggingface/import` | Import from HuggingFace |
| GET | `/v1/huggingface/search` | Search HuggingFace models |
| POST | `/v1/huggingface/inference` | Run inference |

### WebSocket Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/ws` | General WebSocket connection |
| `/v1/ws/training` | Training updates stream |
| `/v1/ws/agent` | Training agent connection |

---

## 💎 Subscription Tiers

| Tier | Price | GPU Access | Training Credits | Features |
|------|-------|------------|------------------|----------|
| **Free** | $0/month | None (local only) | N/A | Local training, Community access, Model publishing |
| **Basic** | $9.99/month | Tesla T4 | 100/month | Cloud training, Priority support, Advanced analytics |
| **Pro** | $29.99/month | Tesla V100 | 500/month | All Basic features, Faster GPUs, More credits |
| **Enterprise** | $99.99/month | A100 | Unlimited | All Pro features, Dedicated support, Custom solutions |

### Training Credits

- 1 credit = 1 minute of training time
- Unused credits roll over monthly
- Additional credits available for purchase
- Free tier users have unlimited local training

---

## 🤗 HuggingFace Integration

AiManage seamlessly integrates with HuggingFace Hub for model portability and sharing.

### Features

- **Push Models**: Upload your trained models to HuggingFace Hub
- **Import Models**: Download and use existing HuggingFace models
- **Search**: Browse the HuggingFace model repository
- **Inference**: Run predictions using HuggingFace's Inference API
- **Model Cards**: Automatic generation of model documentation

### Usage

```python
# Push a model to HuggingFace
POST /v1/huggingface/push
{
  "model_id": "your-model-id",
  "hf_username": "your-hf-username",
  "repo_name": "model-name"
}

# Import a model from HuggingFace
POST /v1/huggingface/import
{
  "repo_id": "username/model-name"
}
```

For detailed setup instructions, see [HUGGINGFACE_SETUP.md](./HUGGINGFACE_SETUP.md).

---

## 📊 Database Schema

### Core Tables

- **users**: User accounts and authentication
- **sessions**: JWT refresh tokens
- **models**: User's training models
- **published_models**: Community marketplace listings
- **model_likes**: User likes on published models
- **model_comments**: Comments on published models
- **model_purchases**: Purchase records
- **model_reviews**: User reviews

### Migrations

Database migrations are managed using golang-migrate. To run migrations:

```bash
cd server

# Migrate up
make migrate-up

# Migrate down
make migrate-down

# Create new migration
make migrate-create name=migration_name
```

---

## 🤝 Contributing

We welcome contributions from the community! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Commit your changes**
   ```bash
   git commit -m 'Add some amazing feature'
   ```
4. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
5. **Open a Pull Request**

### Development Guidelines

- Follow Go best practices for backend code
- Use TypeScript for all frontend code
- Write meaningful commit messages
- Add tests for new features
- Update documentation as needed

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Go Chi Router](https://github.com/go-chi/chi) - Lightweight HTTP router
- [shadcn/ui](https://ui.shadcn.com/) - Beautiful UI components
- [HuggingFace](https://huggingface.co/) - ML model hub and transformers
- [Stripe](https://stripe.com/) - Payment processing
- [PostgreSQL](https://www.postgresql.org/) - Reliable database

---

## 📞 Support

- **Documentation**: See the `/docs` folder for detailed guides
- **Issues**: Open an issue on GitHub
- **Email**: support@aimanage.example.com

---

## 🗺️ Roadmap

- [ ] Multi-GPU training support
- [ ] Advanced model versioning
- [ ] Automated hyperparameter tuning
- [ ] Model deployment API
- [ ] Mobile app (iOS/Android)
- [ ] Collaborative training
- [ ] Model performance benchmarking
- [ ] Custom training environments

---

<div align="center">

**Built with ❤️ by the AiManage Team**

[Website](https://aimanage.example.com) • [Documentation](./docs) • [Report Bug](https://github.com/yourusername/AiManage/issues) • [Request Feature](https://github.com/yourusername/AiManage/issues)

</div>
