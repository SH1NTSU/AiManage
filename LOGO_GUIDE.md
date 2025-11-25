# aiManage Logo Guide

This guide explains the different logo variations available for the aiManage application.

## Logo Concept

The aiManage logo represents an **AI neural network** with a central node connected to surrounding nodes, symbolizing:
- **Artificial Intelligence** - The interconnected neural network structure
- **Management & Control** - The central node as the control point
- **Connectivity** - Lines connecting nodes representing data flow and relationships

## Color Scheme

- **Primary Gradient**: Indigo to Violet (`#6366f1` → `#8b5cf6`)
- **Accent**: Light purple variations for depth
- **Contrast**: White nodes and connections for clarity

## Logo Files

### 1. **favicon.svg** (32x32)
- **Location**: `/app/public/favicon.svg`
- **Usage**: Browser favicon, tab icon
- **Format**: Small, optimized for 16x16 and 32x32 display

### 2. **logo-icon.svg** (64x64)
- **Location**: `/app/public/logo-icon.svg`
- **Usage**: App icons, small UI elements, mobile app icon
- **Format**: Rounded square with padding, works well at small sizes

### 3. **logo.svg** (200x200)
- **Location**: `/app/public/logo.svg`
- **Usage**: Circular logo for profiles, avatars, or splash screens
- **Format**: Full circular design with gradient background

### 4. **logo-horizontal.svg** (400x100)
- **Location**: `/app/public/logo-horizontal.svg`
- **Usage**: Website header, navigation bar, email signatures
- **Format**: Icon + "aiManage" text (ai in gradient, Manage in gray)

## Usage Guidelines

### Where to Use Each Variant

#### Favicon
```html
<link rel="icon" type="image/svg+xml" href="/favicon.svg" />
```

#### Navigation/Header
Use `logo-horizontal.svg` for:
- Website header
- Navigation bar
- Email footers

#### App Icon
Use `logo-icon.svg` for:
- Mobile app icons
- Desktop app icons
- Social media profile pictures

#### Full Logo
Use `logo.svg` for:
- Splash screens
- About pages
- Large promotional materials

## Design Notes

- **Scalable**: All logos are SVG format and scale perfectly
- **Gradient**: Uses CSS gradients that work across all modern browsers
- **Accessibility**: White on colored background provides high contrast
- **Consistent**: All variations use the same neural network motif

## Customization

To change the color scheme, update the gradient stops in each SVG file:

```svg
<linearGradient id="gradient">
  <stop offset="0%" style="stop-color:#6366f1" />   <!-- Change this -->
  <stop offset="100%" style="stop-color:#8b5cf6" /> <!-- And this -->
</linearGradient>
```

## Future Enhancements

Consider creating:
- PNG exports for platforms that don't support SVG
- Dark mode variants
- Monochrome versions for print
- Animation variants for loading screens
