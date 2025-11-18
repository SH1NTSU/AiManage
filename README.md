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

<div align="center">

**Built with ❤️ by the AiManage Team**

[Website](https://aimanage.example.com) • [Documentation](./docs) • [Report Bug](https://github.com/yourusername/AiManage/issues) • [Request Feature](https://github.com/yourusername/AiManage/issues)

</div>
