import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Search, Filter, Download, Star, Eye, DollarSign, Cpu, ArrowRight } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

const Community = () => {
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [priceFilter, setPriceFilter] = useState("all");
  const [sortBy, setSortBy] = useState("popular");
  const [publishedModels, setPublishedModels] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);


// Apply all filters and sorting
  const getFilteredAndSortedModels = () => {
    let filtered = [...publishedModels];

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = filtered.filter(model =>
        model.name.toLowerCase().includes(query) ||
        model.description?.toLowerCase().includes(query) ||
        model.short_description?.toLowerCase().includes(query) ||
        model.tags?.some((tag: string) => tag.toLowerCase().includes(query)) ||
        model.publisher_username?.toLowerCase().includes(query)
      );
    }

    // Category filter
    if (selectedCategory !== "all" && selectedCategory !== "all-categories") {
      filtered = filtered.filter(model => {
        const modelCategory = model.category?.toLowerCase().replace(" ", "-");
        return modelCategory === selectedCategory;
      });
    }

    // Price filter
    if (priceFilter === "free") {
      filtered = filtered.filter(model => model.price === 0);
    } else if (priceFilter === "paid") {
      filtered = filtered.filter(model => model.price > 0);
    }

    // Sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case "popular":
          // Sort by views + downloads combined
          return (b.views_count + b.downloads_count) - (a.views_count + a.downloads_count);

        case "recent":
          // Sort by published date (newest first)
          return new Date(b.published_at).getTime() - new Date(a.published_at).getTime();

        case "rating":
          // Sort by rating average (highest first)
          return (b.rating_average || 0) - (a.rating_average || 0);

        case "downloads":
          // Sort by download count (most first)
          return (b.downloads_count || 0) - (a.downloads_count || 0);

        case "price-low":
          // Sort by price (low to high)
          return (a.price || 0) - (b.price || 0);

        case "price-high":
          // Sort by price (high to low)
          return (b.price || 0) - (a.price || 0);

        default:
          return 0;
      }
    });

    return filtered;
  };

  const filteredModels = getFilteredAndSortedModels();

  useEffect(() => {
    fetchPublishedModels();
  }, []);

  const fetchPublishedModels = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        console.error("No token found");
        setLoading(false);
        return;
      }

      const response = await fetch("http://localhost:8081/v1/published-models", {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to fetch published models");
      }

      const data = await response.json();
      setPublishedModels(data || []);
    } catch (error) {
      console.error("Error fetching published models:", error);
    } finally {
      setLoading(false);
    }
  };

  const categories = [
    "All Categories",
    "Image Classification",
    "Object Detection",
    "NLP",
    "Computer Vision",
    "Time Series",
    "Audio Processing",
  ];

  const formatPrice = (priceInCents: number) => {
    if (priceInCents === 0) return "Free";
    return `$${(priceInCents / 100).toFixed(2)}`;
  };

  return (
    <div className="space-y-6 animate-slide-up">
      {/* Header */}
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Community Marketplace</h2>
        <p className="text-muted-foreground mt-1">
          Discover and download AI models from the community
        </p>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-col gap-4">
        {/* Search Bar */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search models by name, description, or tags..."
            className="pl-10"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        {/* Filters Row */}
        <div className="flex flex-wrap gap-4">
          {/* Category Filter */}
          <select
            className="px-4 py-2 rounded-md border border-border bg-background text-foreground"
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
          >
            {categories.map((cat) => (
              <option key={cat} value={cat.toLowerCase().replace(" ", "-")}>
                {cat}
              </option>
            ))}
          </select>

          {/* Price Filter */}
          <select
            className="px-4 py-2 rounded-md border border-border bg-background text-foreground"
            value={priceFilter}
            onChange={(e) => setPriceFilter(e.target.value)}
          >
            <option value="all">All Prices</option>
            <option value="free">Free Only</option>
            <option value="paid">Paid Only</option>
          </select>

          {/* Sort By */}
          <select
            className="px-4 py-2 rounded-md border border-border bg-background text-foreground"
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
          >
            <option value="popular">Most Popular</option>
            <option value="recent">Recently Added</option>
            <option value="rating">Highest Rated</option>
            <option value="downloads">Most Downloads</option>
            <option value="price-low">Price: Low to High</option>
            <option value="price-high">Price: High to Low</option>
          </select>
        </div>
      </div>

      {/* Results count */}
      {publishedModels.length > 0 && (
        <div className="text-sm text-muted-foreground">
          Showing <span className="font-semibold text-foreground">{filteredModels.length}</span> of <span className="font-semibold text-foreground">{publishedModels.length}</span> models
        </div>
      )}

      {/* Models Grid */}
      {publishedModels.length === 0 ? (
        <Card className="bg-gradient-card border-border shadow-card">
          <CardContent className="flex flex-col items-center justify-center py-16">
            <Cpu className="w-20 h-20 text-muted-foreground mb-4" />
            <h3 className="text-2xl font-semibold mb-2">No Models Yet</h3>
            <p className="text-muted-foreground text-center max-w-md">
              The community marketplace is empty. Be the first to publish your trained model!
            </p>
            <Button className="mt-6" onClick={() => window.location.href = "/"}>
              Go to My Models
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredModels.map((model) => (
            <Card
              key={model.id}
              className="bg-gradient-card border-border hover:border-primary/50 transition-all shadow-card hover:shadow-glow group cursor-pointer"
              onClick={() => navigate(`/community/${model.id}`)}
            >
              <CardHeader>
                <div className="flex items-start justify-between">
                  {model.picture ? (
                    <div className="w-16 h-16 rounded-xl overflow-hidden shadow-lg ring-2 ring-primary/20 group-hover:ring-primary/40 transition-all">
                      <img
                        src={`http://localhost:8081${model.picture}`}
                        alt={model.name}
                        className="w-full h-full object-cover"
                      />
                    </div>
                  ) : (
                    <div className="w-16 h-16 rounded-xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center group-hover:from-primary/30 group-hover:to-secondary/30 transition-all shadow-lg">
                      <Cpu className="w-8 h-8 text-primary" />
                    </div>
                  )}

                  <div className="flex flex-col gap-2">
                    <Badge
                      className={`${
                        model.price === 0
                          ? "bg-green-500/20 text-green-500 border-green-500/30"
                          : "bg-primary/20 text-primary border-primary/30"
                      }`}
                    >
                      {formatPrice(model.price)}
                    </Badge>
                    {model.is_featured && (
                      <Badge className="bg-yellow-500/20 text-yellow-500 border-yellow-500/30">
                        Featured
                      </Badge>
                    )}
                  </div>
                </div>

                <CardTitle className="mt-4">{model.name}</CardTitle>
                {model.publisher_username && (
                  <p className="text-sm text-muted-foreground mt-1">
                    by <span className="font-medium text-primary">{model.publisher_username}</span>
                  </p>
                )}
                <CardDescription className="line-clamp-2 mt-2">
                  {model.short_description || model.description}
                </CardDescription>

                {/* Tags */}
                {model.tags && model.tags.length > 0 && (
                  <div className="flex flex-wrap gap-1 mt-2">
                    {model.tags.slice(0, 3).map((tag: string, idx: number) => (
                      <Badge key={idx} variant="outline" className="text-xs">
                        {tag}
                      </Badge>
                    ))}
                    {model.tags.length > 3 && (
                      <Badge variant="outline" className="text-xs">
                        +{model.tags.length - 3}
                      </Badge>
                    )}
                  </div>
                )}
              </CardHeader>

              <CardContent className="space-y-3">
                {/* Stats */}
                <div className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Star className="w-4 h-4 text-yellow-500 fill-yellow-500" />
                    <span>{model.rating_average?.toFixed(1) || "0.0"}</span>
                    <span className="text-xs">({model.rating_count || 0})</span>
                  </div>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Download className="w-4 h-4" />
                    <span>{model.downloads_count || 0}</span>
                  </div>
                  <div className="flex items-center gap-1 text-muted-foreground">
                    <Eye className="w-4 h-4" />
                    <span>{model.views_count || 0}</span>
                  </div>
                </div>

                {/* Metadata */}
                {(model.framework || model.model_type) && (
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    {model.framework && (
                      <Badge variant="secondary" className="text-xs">
                        {model.framework}
                      </Badge>
                    )}
                    {model.model_type && (
                      <Badge variant="secondary" className="text-xs">
                        {model.model_type}
                      </Badge>
                    )}
                  </div>
                )}

                {/* Action Button */}
                <Button
                  className="w-full bg-gradient-primary hover:opacity-90"
                  size="sm"
                  onClick={(e) => {
                    e.stopPropagation();
                    navigate(`/community/${model.id}`);
                  }}
                >
                  View Details
                  <ArrowRight className="w-4 h-4 ml-2" />
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
};

export default Community;
