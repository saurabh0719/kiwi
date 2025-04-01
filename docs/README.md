# Kiwi Landing Page

This directory contains the source files for the Kiwi landing page hosted on GitHub Pages.

## Structure

- `index.html` - Main landing page HTML
- `styles.css` - CSS styles for the landing page
- `script.js` - JavaScript for interactive elements
- `_config.yml` - GitHub Pages configuration
- `CNAME` - Optional custom domain configuration

## Local Development

To develop and test locally:

1. Install Jekyll (if you want to test with GitHub Pages): 
   ```
   gem install jekyll bundler
   ```

2. Run a local server:
   ```
   cd docs
   jekyll serve
   ```

   Or simply use any static file server, like Python's built-in server:
   ```
   cd docs
   python3 -m http.server
   ```

3. Open your browser to `http://localhost:8000` or `http://localhost:4000` (Jekyll)

## Deployment

The site will be automatically deployed to GitHub Pages when:

1. The GitHub repository is configured to use the `/docs` folder for GitHub Pages
2. Changes are pushed to the main branch

## Configuration

To configure GitHub Pages to use this directory:

1. Go to your GitHub repository
2. Click "Settings"
3. Scroll down to "GitHub Pages"
4. Under "Source", select "main branch /docs folder"
5. Click "Save"

Your site will be available at `https://username.github.io/kiwi/` or your custom domain if configured. 