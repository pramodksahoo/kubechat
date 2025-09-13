import React, { useState, useEffect, useRef, useCallback, ReactNode } from 'react';
import { Button, Heading, Body, Input, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
// Removed unused import
import { ScreenReaderUtils } from '../../design-system/accessibility';
import { trackEvent } from '../../lib/monitoring';

// Documentation article structure
export interface HelpArticle {
  id: string;
  title: string;
  content: ReactNode;
  category: string;
  tags: string[];
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  lastUpdated: Date;
  author?: string;
  relatedArticles?: string[];
  searchKeywords: string[];
  estimatedReadTime?: number; // in minutes
}

// Documentation section
export interface HelpSection {
  id: string;
  title: string;
  description: string;
  icon: ReactNode;
  articles: HelpArticle[];
  order: number;
}

// Search result
export interface SearchResult {
  article: HelpArticle;
  relevance: number;
  matchedKeywords: string[];
  excerpt: string;
}

// Help center component
export interface HelpCenterProps {
  sections: HelpSection[];
  featuredArticles?: string[];
  onArticleView?: (articleId: string) => void;
  onSearch?: (query: string, results: SearchResult[]) => void;
  className?: string;
}

export function HelpCenter({
  sections,
  featuredArticles = [],
  onArticleView,
  onSearch,
  className,
}: HelpCenterProps) {
  const [selectedArticle, setSelectedArticle] = useState<HelpArticle | null>(null);
  const [selectedSection, setSelectedSection] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);

  const sidebarRef = useRef<HTMLDivElement>(null);
  const contentRef = useRef<HTMLDivElement>(null);

  // Flatten all articles for search
  const allArticles = sections.flatMap(section => section.articles);

  // Featured articles
  const featured = featuredArticles
    .map(id => allArticles.find(article => article.id === id))
    .filter(Boolean) as HelpArticle[];

  // Search functionality
  const performSearch = useCallback((query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      setIsSearching(false);
      return;
    }

    setIsSearching(true);
    
    const results: SearchResult[] = [];
    const queryWords = query.toLowerCase().split(' ').filter(word => word.length > 2);

    allArticles.forEach(article => {
      let relevance = 0;
      const matchedKeywords: string[] = [];
      
      // Search in title (higher weight)
      queryWords.forEach(word => {
        if (article.title.toLowerCase().includes(word)) {
          relevance += 10;
          matchedKeywords.push(word);
        }
      });

      // Search in tags (medium weight)
      article.tags.forEach(tag => {
        queryWords.forEach(word => {
          if (tag.toLowerCase().includes(word)) {
            relevance += 5;
            matchedKeywords.push(word);
          }
        });
      });

      // Search in keywords (medium weight)
      article.searchKeywords.forEach(keyword => {
        queryWords.forEach(word => {
          if (keyword.toLowerCase().includes(word)) {
            relevance += 5;
            matchedKeywords.push(word);
          }
        });
      });

      // Search in content (lower weight)
      const contentText = extractTextFromReactNode(article.content);
      queryWords.forEach(word => {
        const matches = (contentText.toLowerCase().match(new RegExp(word, 'g')) || []).length;
        relevance += matches * 2;
        if (matches > 0) {
          matchedKeywords.push(word);
        }
      });

      if (relevance > 0) {
        const excerpt = generateExcerpt(contentText, queryWords[0]);
        results.push({
          article,
          relevance,
          matchedKeywords: [...new Set(matchedKeywords)],
          excerpt,
        });
      }
    });

    // Sort by relevance
    results.sort((a, b) => b.relevance - a.relevance);
    
    setSearchResults(results);
    setIsSearching(false);
    onSearch?.(query, results);
    
    trackEvent('help_search', {
      query,
      resultsCount: results.length,
    });
  }, [allArticles, onSearch]);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      performSearch(searchQuery);
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery, performSearch]);

  // Handle article selection
  const handleArticleSelect = (article: HelpArticle) => {
    setSelectedArticle(article);
    setSelectedSection(null);
    onArticleView?.(article.id);
    
    trackEvent('help_article_viewed', {
      articleId: article.id,
      category: article.category,
      difficulty: article.difficulty,
    });

    // Announce to screen readers
    ScreenReaderUtils.announce(`Viewing help article: ${article.title}`);
  };

  // Handle section selection
  const handleSectionSelect = (sectionId: string) => {
    setSelectedSection(sectionId);
    setSelectedArticle(null);
  };

  // Handle back navigation
  const handleBack = () => {
    if (selectedArticle) {
      setSelectedArticle(null);
    } else if (selectedSection) {
      setSelectedSection(null);
    }
  };

  // Get breadcrumb
  const getBreadcrumb = () => {
    const crumbs = ['Help Center'];
    
    if (selectedSection) {
      const section = sections.find(s => s.id === selectedSection);
      if (section) crumbs.push(section.title);
    }
    
    if (selectedArticle) {
      crumbs.push(selectedArticle.title);
    }
    
    return crumbs;
  };

  return (
    <div className={cn('flex h-full bg-background-primary', className)}>
      {/* Sidebar */}
      <div 
        ref={sidebarRef}
        className="w-80 bg-background-secondary border-r border-border-primary flex flex-col"
      >
        {/* Search */}
        <div className="p-4 border-b border-border-primary">
          <Input
            type="search"
            placeholder="Search help articles..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            leftIcon={
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="m21 21-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            }
          />
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto scrollbar-thin">
          {searchQuery ? (
            /* Search Results */
            <div className="p-4">
              <Heading level={4} className="mb-4">
                {isSearching ? 'Searching...' : `${searchResults.length} results`}
              </Heading>
              
              {searchResults.map((result) => (
                <button
                  key={result.article.id}
                  onClick={() => handleArticleSelect(result.article)}
                  className="w-full text-left p-3 rounded-lg hover:bg-background-tertiary transition-colors mb-2"
                >
                  <Body className="font-medium mb-1">{result.article.title}</Body>
                  <Body size="sm" color="secondary" className="mb-2 line-clamp-2">
                    {result.excerpt}
                  </Body>
                  <div className="flex items-center gap-2">
                    <StatusBadge variant="neutral">{result.article.category}</StatusBadge>
                    <StatusBadge variant="info">{result.article.difficulty}</StatusBadge>
                  </div>
                </button>
              ))}
              
              {!isSearching && searchResults.length === 0 && (
                <Body color="tertiary">No articles found for &quot;{searchQuery}&quot;</Body>
              )}
            </div>
          ) : selectedSection ? (
            /* Section Articles */
            <div className="p-4">
              <Button variant="ghost" size="sm" onClick={handleBack} className="mb-4">
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
                Back
              </Button>
              
              {sections
                .find(s => s.id === selectedSection)
                ?.articles.map((article) => (
                  <button
                    key={article.id}
                    onClick={() => handleArticleSelect(article)}
                    className="w-full text-left p-3 rounded-lg hover:bg-background-tertiary transition-colors mb-2"
                  >
                    <Body className="font-medium mb-1">{article.title}</Body>
                    <div className="flex items-center gap-2">
                      <StatusBadge variant="info">{article.difficulty}</StatusBadge>
                      {article.estimatedReadTime && (
                        <Body size="sm" color="tertiary">
                          {article.estimatedReadTime} min read
                        </Body>
                      )}
                    </div>
                  </button>
                ))}
            </div>
          ) : (
            /* Sections and Featured */
            <div>
              {/* Featured Articles */}
              {featured.length > 0 && (
                <div className="p-4 border-b border-border-primary">
                  <Heading level={4} className="mb-3">Featured</Heading>
                  {featured.map((article) => (
                    <button
                      key={article.id}
                      onClick={() => handleArticleSelect(article)}
                      className="w-full text-left p-3 rounded-lg hover:bg-background-tertiary transition-colors mb-2"
                    >
                      <Body className="font-medium mb-1">{article.title}</Body>
                      <Body size="sm" color="secondary">
                        {article.category}
                      </Body>
                    </button>
                  ))}
                </div>
              )}

              {/* Sections */}
              <div className="p-4">
                <Heading level={4} className="mb-3">Browse by Category</Heading>
                {sections
                  .sort((a, b) => a.order - b.order)
                  .map((section) => (
                    <button
                      key={section.id}
                      onClick={() => handleSectionSelect(section.id)}
                      className="w-full text-left p-3 rounded-lg hover:bg-background-tertiary transition-colors mb-2 flex items-start gap-3"
                    >
                      <div className="w-6 h-6 mt-0.5 text-text-secondary">
                        {section.icon}
                      </div>
                      <div className="flex-1">
                        <Body className="font-medium mb-1">{section.title}</Body>
                        <Body size="sm" color="secondary" className="mb-2">
                          {section.description}
                        </Body>
                        <Body size="sm" color="tertiary">
                          {section.articles.length} articles
                        </Body>
                      </div>
                      <svg className="w-4 h-4 mt-1 text-text-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                    </button>
                  ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div ref={contentRef} className="flex-1 flex flex-col overflow-hidden">
        {/* Breadcrumb */}
        <div className="p-4 border-b border-border-primary bg-background-secondary">
          <nav aria-label="Breadcrumb">
            <ol className="flex items-center space-x-2">
              {getBreadcrumb().map((crumb, index, array) => (
                <li key={index} className="flex items-center">
                  {index > 0 && (
                    <svg className="w-4 h-4 mx-2 text-text-tertiary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  )}
                  <Body 
                    size="sm" 
                    color={index === array.length - 1 ? 'primary' : 'secondary'}
                    className={index < array.length - 1 ? 'hover:text-text-primary cursor-pointer' : ''}
                  >
                    {crumb}
                  </Body>
                </li>
              ))}
            </ol>
          </nav>
        </div>

        {/* Content Area */}
        <div className="flex-1 overflow-y-auto scrollbar-thin">
          {selectedArticle ? (
            <ArticleView article={selectedArticle} onBack={handleBack} />
          ) : (
            <div className="p-8 text-center">
              <div className="max-w-md mx-auto">
                <div className="w-16 h-16 mx-auto mb-4 text-text-tertiary">
                  <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
                    <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-8-3a1 1 0 00-.867.5 1 1 0 11-1.731-1A3 3 0 0113 8a3.001 3.001 0 01-2 2.83V11a1 1 0 11-2 0v-1a1 1 0 011-1 1 1 0 100-2zm0 8a1 1 0 100-2 1 1 0 000 2z" clipRule="evenodd" />
                  </svg>
                </div>
                <Heading level={3} className="mb-2">KubeChat Help Center</Heading>
                <Body color="secondary" className="mb-6">
                  Search for help articles or browse by category to get started.
                </Body>
                <Button variant="primary" onClick={() => setSearchQuery('getting started')}>
                  Getting Started Guide
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// Article view component
function ArticleView({ 
  article, 
  onBack 
}: { 
  article: HelpArticle;
  onBack: () => void;
}) {
  const [helpful, setHelpful] = useState<boolean | null>(null);

  const handleFeedback = (isHelpful: boolean) => {
    setHelpful(isHelpful);
    trackEvent('help_article_feedback', {
      articleId: article.id,
      helpful: isHelpful,
    });
  };

  return (
    <div className="max-w-4xl mx-auto p-8">
      {/* Header */}
      <div className="mb-8">
        <Button variant="ghost" size="sm" onClick={onBack} className="mb-4">
          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back to Help Center
        </Button>
        
        <div className="flex items-start justify-between mb-4">
          <div>
            <Heading level={1} className="mb-2">{article.title}</Heading>
            <div className="flex items-center gap-4">
              <StatusBadge variant="neutral">{article.category}</StatusBadge>
              <StatusBadge variant="info">{article.difficulty}</StatusBadge>
              {article.estimatedReadTime && (
                <Body size="sm" color="tertiary">
                  {article.estimatedReadTime} minute read
                </Body>
              )}
            </div>
          </div>
        </div>

        <div className="flex items-center gap-4 text-sm text-text-tertiary">
          {article.author && <span>By {article.author}</span>}
          <span>Updated {article.lastUpdated.toLocaleDateString()}</span>
        </div>
      </div>

      {/* Content */}
      <div className="prose prose-gray dark:prose-invert max-w-none mb-8">
        {article.content}
      </div>

      {/* Tags */}
      {article.tags.length > 0 && (
        <div className="mb-8">
          <Body size="sm" className="font-medium mb-2">Tags:</Body>
          <div className="flex flex-wrap gap-2">
            {article.tags.map((tag) => (
              <span
                key={tag}
                className="px-2 py-1 text-xs bg-background-secondary rounded-full"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>
      )}

      {/* Feedback */}
      <div className="border-t border-border-primary pt-8">
        <Heading level={4} className="mb-4">Was this article helpful?</Heading>
        
        {helpful === null ? (
          <div className="flex gap-3">
            <Button variant="outline" onClick={() => handleFeedback(true)}>
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 10h4.764a2 2 0 011.789 2.894l-3.5 7A2 2 0 0115.263 21h-4.017c-.163 0-.326-.02-.485-.06L7 20m7-10V5a2 2 0 00-2-2h-.095c-.5 0-.905.405-.905.905 0 .714-.211 1.412-.608 2.006L7 11v9m7-10h-2M7 20H5a2 2 0 01-2-2v-6a2 2 0 012-2h2.5" />
              </svg>
              Yes, this was helpful
            </Button>
            <Button variant="outline" onClick={() => handleFeedback(false)}>
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14H5.236a2 2 0 01-1.789-2.894l3.5-7A2 2 0 018.736 3h4.018c.163 0 .326.02.485.06L17 4m-7 10v2a2 2 0 002 2h.095c.5 0 .905-.405.905-.905 0-.714.211-1.412.608-2.006L17 9v-5m-7 10h2m5-10H9a2 2 0 00-2 2v6a2 2 0 002 2h2.5" />
              </svg>
              No, this wasn&apos;t helpful
            </Button>
          </div>
        ) : (
          <div className="p-4 bg-background-secondary rounded-lg">
            <Body>
              {helpful 
                ? "Thank you for your feedback! We&apos;re glad this article was helpful."
                : "Thank you for your feedback. We&apos;ll work on improving this article."
              }
            </Body>
          </div>
        )}
      </div>
    </div>
  );
}

// Utility functions
function extractTextFromReactNode(node: ReactNode): string {
  if (typeof node === 'string') return node;
  if (typeof node === 'number') return String(node);
  if (React.isValidElement(node)) {
    if (node.props.children) {
      return extractTextFromReactNode(node.props.children);
    }
  }
  if (Array.isArray(node)) {
    return node.map(extractTextFromReactNode).join(' ');
  }
  return '';
}

function generateExcerpt(text: string, searchTerm: string): string {
  const index = text.toLowerCase().indexOf(searchTerm.toLowerCase());
  if (index === -1) return text.substring(0, 150) + '...';
  
  const start = Math.max(0, index - 75);
  const end = Math.min(text.length, index + 75);
  
  return (start > 0 ? '...' : '') + 
         text.substring(start, end) + 
         (end < text.length ? '...' : '');
}