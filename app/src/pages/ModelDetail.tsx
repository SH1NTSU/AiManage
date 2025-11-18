import { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { jwtDecode } from "jwt-decode";
import {
  Download,
  Star,
  Eye,
  DollarSign,
  Calendar,
  User,
  Package,
  Code,
  ArrowLeft,
  Cpu,
  Shield,
  TrendingUp,
  Heart,
  MessageCircle,
  Send,
  Trash2
} from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import StripeProvider from "@/components/StripeProvider";
import StripeCheckout from "@/components/StripeCheckout";

interface PublishedModel {
  id: number;
  model_id: number;
  publisher_id: number;
  publisher_username: string;
  name: string;
  picture: string;
  description: string;
  short_description: string;
  price: number;
  category: string;
  tags: string[];
  model_type: string;
  framework: string;
  file_size: number;
  accuracy_score: number;
  license_type: string;
  downloads_count: number;
  views_count: number;
  rating_average: number;
  rating_count: number;
  is_featured: boolean;
  published_at: string;
  updated_at: string;
  trained_model_path: string;
}

interface Comment {
  id: number;
  user_id: number;
  username: string;
  comment_text: string;
  created_at: string;
  updated_at: string;
  edited: boolean;
  parent_comment_id: number | null;
}

interface JWTPayload {
  userID: string;
  email: string;
  exp: number;
}

const ModelDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { toast } = useToast();
  const [model, setModel] = useState<PublishedModel | null>(null);
  const [loading, setLoading] = useState(true);
  const [downloading, setDownloading] = useState(false);
  const [checkoutOpen, setCheckoutOpen] = useState(false);

  // Likes state
  const [liked, setLiked] = useState(false);
  const [likesCount, setLikesCount] = useState(0);
  const [likingInProgress, setLikingInProgress] = useState(false);

  // Comments state
  const [comments, setComments] = useState<Comment[]>([]);
  const [newComment, setNewComment] = useState("");
  const [submittingComment, setSubmittingComment] = useState(false);
  const [loadingComments, setLoadingComments] = useState(true);

  useEffect(() => {
    fetchModelDetails();
    fetchLikes();
    fetchComments();
  }, [id]);

  const fetchModelDetails = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication required",
          description: "Please login to view model details",
          variant: "destructive",
        });
        navigate("/auth");
        return;
      }

      const response = await fetch(`http://localhost:8081/v1/published-models/${id}`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to fetch model details");
      }

      const data = await response.json();
      setModel(data);
    } catch (error) {
      console.error("Error fetching model details:", error);
      toast({
        title: "Error",
        description: "Failed to load model details",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  };

  const fetchLikes = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return;

      const response = await fetch(`http://localhost:8081/v1/published-models/${id}/likes`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setLikesCount(data.likes_count || 0);
        setLiked(data.user_liked || false);
      }
    } catch (error) {
      console.error("Error fetching likes:", error);
    }
  };

  const fetchComments = async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return;

      const response = await fetch(`http://localhost:8081/v1/published-models/${id}/comments`, {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setComments(data || []);
      }
    } catch (error) {
      console.error("Error fetching comments:", error);
    } finally {
      setLoadingComments(false);
    }
  };

  const handleLikeToggle = async () => {
    if (likingInProgress) return;

    setLikingInProgress(true);
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication required",
          description: "Please login to like models",
          variant: "destructive",
        });
        navigate("/auth");
        return;
      }

      const method = liked ? "DELETE" : "POST";
      const response = await fetch(`http://localhost:8081/v1/published-models/${id}/like`, {
        method,
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setLiked(!liked);
        setLikesCount(data.likes_count || 0);
        toast({
          title: liked ? "Unliked" : "Liked",
          description: liked ? "Removed from your liked models" : "Added to your liked models",
        });
      }
    } catch (error) {
      console.error("Error toggling like:", error);
      toast({
        title: "Error",
        description: "Failed to update like status",
        variant: "destructive",
      });
    } finally {
      setLikingInProgress(false);
    }
  };

  const handleAddComment = async () => {
    if (!newComment.trim() || submittingComment) return;

    setSubmittingComment(true);
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication required",
          description: "Please login to comment",
          variant: "destructive",
        });
        navigate("/auth");
        return;
      }

      const response = await fetch(`http://localhost:8081/v1/published-models/${id}/comments`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          comment_text: newComment,
        }),
      });

      if (response.ok) {
        setNewComment("");
        fetchComments(); // Refresh comments
        toast({
          title: "Comment added",
          description: "Your comment has been posted successfully",
        });
      } else {
        throw new Error("Failed to add comment");
      }
    } catch (error) {
      console.error("Error adding comment:", error);
      toast({
        title: "Error",
        description: "Failed to post comment",
        variant: "destructive",
      });
    } finally {
      setSubmittingComment(false);
    }
  };

  const handleDeleteComment = async (commentId: number) => {
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication required",
          variant: "destructive",
        });
        return;
      }

      const response = await fetch(`http://localhost:8081/v1/comments/${commentId}`, {
        method: "DELETE",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        fetchComments(); // Refresh comments
        toast({
          title: "Comment deleted",
          description: "Your comment has been removed",
        });
      } else {
        throw new Error("Failed to delete comment");
      }
    } catch (error) {
      console.error("Error deleting comment:", error);
      toast({
        title: "Error",
        description: "Failed to delete comment",
        variant: "destructive",
      });
    }
  };

  const handleDownloadClick = () => {
    if (!model) return;

    // Check if model is paid
    if (model.price > 0) {
      // Open checkout for paid models
      setCheckoutOpen(true);
    } else {
      // Direct download for free models
      handleDownload();
    }
  };

  const handleDownload = async () => {
    if (!model) return;

    setDownloading(true);
    try {
      const token = localStorage.getItem("token");
      if (!token) {
        toast({
          title: "Authentication required",
          description: "Please login to download models",
          variant: "destructive",
        });
        navigate("/auth");
        return;
      }

      const response = await fetch(`http://localhost:8081/v1/published-models/${id}/download`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error("Failed to download model");
      }

      // Get filename from Content-Disposition header
      const contentDisposition = response.headers.get("Content-Disposition");
      let filename = `${model.name}.zip`;
      if (contentDisposition) {
        const filenameMatch = contentDisposition.match(/filename="(.+)"/);
        if (filenameMatch) {
          filename = filenameMatch[1];
        }
      }

      // Download the file
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      toast({
        title: "Download started",
        description: `Downloading ${filename}`,
      });

      // Refresh model details to update download count
      fetchModelDetails();
    } catch (error) {
      console.error("Error downloading model:", error);
      toast({
        title: "Download failed",
        description: "Failed to download the model. Please try again.",
        variant: "destructive",
      });
    } finally {
      setDownloading(false);
    }
  };

  const formatPrice = (priceInCents: number) => {
    if (priceInCents === 0) return "Free";
    return `$${(priceInCents / 100).toFixed(2)}`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  };

  const formatFileSize = (bytes: number) => {
    if (!bytes) return "N/A";
    const mb = bytes / (1024 * 1024);
    if (mb < 1024) {
      return `${mb.toFixed(2)} MB`;
    }
    return `${(mb / 1024).toFixed(2)} GB`;
  };

  const getCurrentUserId = (): number | null => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return null;
      const decoded = jwtDecode<JWTPayload>(token);
      return parseInt(decoded.userID);
    } catch (error) {
      console.error("Error decoding token:", error);
      return null;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <Cpu className="w-16 h-16 text-primary animate-pulse mx-auto mb-4" />
          <p className="text-muted-foreground">Loading model details...</p>
        </div>
      </div>
    );
  }

  if (!model) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Card className="max-w-md">
          <CardContent className="pt-6 text-center">
            <h3 className="text-2xl font-semibold mb-2">Model Not Found</h3>
            <p className="text-muted-foreground mb-4">
              The model you're looking for doesn't exist or has been removed.
            </p>
            <Button onClick={() => navigate("/community")}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back to Community
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6 animate-slide-up pb-8">
      {/* Back Button */}
      <Button
        variant="ghost"
        onClick={() => navigate("/community")}
        className="mb-4"
      >
        <ArrowLeft className="w-4 h-4 mr-2" />
        Back to Community
      </Button>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - Main Details */}
        <div className="lg:col-span-2 space-y-6">
          {/* Header Card */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <div className="flex flex-col md:flex-row gap-6">
                {/* Model Image */}
                {model.picture ? (
                  <div className="w-32 h-32 rounded-xl overflow-hidden shadow-lg ring-2 ring-primary/20">
                    <img
                      src={`http://localhost:8081${model.picture}`}
                      alt={model.name}
                      className="w-full h-full object-cover"
                    />
                  </div>
                ) : (
                  <div className="w-32 h-32 rounded-xl bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center shadow-lg">
                    <Cpu className="w-16 h-16 text-primary" />
                  </div>
                )}

                {/* Title and Meta */}
                <div className="flex-1">
                  <div className="flex flex-wrap gap-2 mb-3">
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
                    {model.category && (
                      <Badge variant="outline">{model.category}</Badge>
                    )}
                  </div>

                  <CardTitle className="text-3xl mb-2">{model.name}</CardTitle>

                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1">
                      <User className="w-4 h-4" />
                      <span className="font-medium text-primary">
                        {model.publisher_username || "Unknown"}
                      </span>
                    </div>
                    <div className="flex items-center gap-1">
                      <Calendar className="w-4 h-4" />
                      <span>{formatDate(model.published_at)}</span>
                    </div>
                  </div>

                  {/* Stats */}
                  <div className="flex items-center gap-6 mt-4">
                    <div className="flex items-center gap-1">
                      <Star className="w-5 h-5 text-yellow-500 fill-yellow-500" />
                      <span className="font-semibold">{model.rating_average?.toFixed(1) || "0.0"}</span>
                      <span className="text-sm text-muted-foreground">({model.rating_count || 0})</span>
                    </div>
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Download className="w-5 h-5" />
                      <span className="font-semibold">{model.downloads_count || 0}</span>
                      <span className="text-sm">downloads</span>
                    </div>
                    <div className="flex items-center gap-1 text-muted-foreground">
                      <Eye className="w-5 h-5" />
                      <span className="font-semibold">{model.views_count || 0}</span>
                      <span className="text-sm">views</span>
                    </div>
                    <button
                      onClick={handleLikeToggle}
                      disabled={likingInProgress}
                      className="flex items-center gap-1 text-muted-foreground hover:text-red-500 transition-colors disabled:opacity-50"
                    >
                      <Heart
                        className={`w-5 h-5 transition-all ${
                          liked ? "fill-red-500 text-red-500" : ""
                        }`}
                      />
                      <span className="font-semibold">{likesCount}</span>
                      <span className="text-sm">likes</span>
                    </button>
                  </div>
                </div>
              </div>
            </CardHeader>
          </Card>

          {/* Description Card */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <CardTitle>Description</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground whitespace-pre-wrap leading-relaxed">
                {model.description || model.short_description || "No description provided."}
              </p>
            </CardContent>
          </Card>

          {/* Tags */}
          {model.tags && model.tags.length > 0 && (
            <Card className="bg-gradient-card border-border shadow-card">
              <CardHeader>
                <CardTitle>Tags</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  {model.tags.map((tag: string, idx: number) => (
                    <Badge key={idx} variant="secondary" className="text-sm">
                      {tag}
                    </Badge>
                  ))}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Comments Section */}
          <Card className="bg-gradient-card border-border shadow-card">
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="flex items-center gap-2">
                  <MessageCircle className="w-5 h-5" />
                  Comments ({comments.length})
                </CardTitle>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              {/* Comment Form */}
              <div className="space-y-3">
                <Textarea
                  placeholder="Share your thoughts about this model..."
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  className="min-h-[100px] resize-none"
                  maxLength={2000}
                />
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">
                    {newComment.length}/2000
                  </span>
                  <Button
                    onClick={handleAddComment}
                    disabled={!newComment.trim() || submittingComment}
                    size="sm"
                  >
                    <Send className="w-4 h-4 mr-2" />
                    {submittingComment ? "Posting..." : "Post Comment"}
                  </Button>
                </div>
              </div>

              <Separator />

              {/* Comments List */}
              <div className="space-y-4">
                {loadingComments ? (
                  <div className="text-center py-8">
                    <MessageCircle className="w-12 h-12 text-muted-foreground animate-pulse mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground">Loading comments...</p>
                  </div>
                ) : comments.length === 0 ? (
                  <div className="text-center py-8">
                    <MessageCircle className="w-12 h-12 text-muted-foreground mx-auto mb-2" />
                    <p className="text-sm text-muted-foreground">
                      No comments yet. Be the first to share your thoughts!
                    </p>
                  </div>
                ) : (
                  comments.map((comment) => {
                    const currentUserId = getCurrentUserId();
                    const isOwner = currentUserId && comment.user_id === currentUserId;

                    return (
                      <div
                        key={comment.id}
                        className="flex gap-3 p-4 rounded-lg bg-card/50 border border-border/50"
                      >
                        <div className="flex-shrink-0">
                          <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center">
                            <User className="w-5 h-5 text-primary" />
                          </div>
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-start justify-between gap-2 mb-2">
                            <div>
                              <p className="font-semibold text-sm">
                                {comment.username || "Anonymous"}
                              </p>
                              <p className="text-xs text-muted-foreground">
                                {formatDate(comment.created_at)}
                                {comment.edited && " (edited)"}
                              </p>
                            </div>
                            {isOwner && (
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => handleDeleteComment(comment.id)}
                                className="text-destructive hover:text-destructive hover:bg-destructive/10"
                              >
                                <Trash2 className="w-4 h-4" />
                              </Button>
                            )}
                          </div>
                          <p className="text-sm text-muted-foreground whitespace-pre-wrap break-words">
                            {comment.comment_text}
                          </p>
                        </div>
                      </div>
                    );
                  })
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Right Column - Actions & Specs */}
        <div className="space-y-6">
          {/* Download Card */}
          <Card className="bg-gradient-card border-border shadow-card sticky top-6">
            <CardHeader>
              <CardTitle className="text-2xl">
                {formatPrice(model.price)}
              </CardTitle>
              <CardDescription>
                {model.price === 0 ? "Free to download" : "One-time purchase"}
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Button
                onClick={handleDownloadClick}
                disabled={downloading}
                className="w-full bg-gradient-primary hover:opacity-90"
                size="lg"
              >
                <Download className="w-5 h-5 mr-2" />
                {downloading ? "Downloading..." : model.price === 0 ? "Download Now" : `Buy for ${formatPrice(model.price)}`}
              </Button>

              <Separator />

              {/* Specifications */}
              <div className="space-y-3">
                <h4 className="font-semibold text-sm text-muted-foreground uppercase">
                  Specifications
                </h4>

                {model.model_type && (
                  <div className="flex items-start gap-3">
                    <Cpu className="w-5 h-5 text-muted-foreground mt-0.5" />
                    <div>
                      <p className="text-sm font-medium">Model Type</p>
                      <p className="text-sm text-muted-foreground">{model.model_type}</p>
                    </div>
                  </div>
                )}

                {model.framework && (
                  <div className="flex items-start gap-3">
                    <Code className="w-5 h-5 text-muted-foreground mt-0.5" />
                    <div>
                      <p className="text-sm font-medium">Framework</p>
                      <p className="text-sm text-muted-foreground">{model.framework}</p>
                    </div>
                  </div>
                )}

                {model.accuracy_score != null && model.accuracy_score !== undefined && (
                  <div className="flex items-start gap-3">
                    <TrendingUp className="w-5 h-5 text-muted-foreground mt-0.5" />
                    <div>
                      <p className="text-sm font-medium">Accuracy Score</p>
                      <p className="text-sm text-muted-foreground">{(model.accuracy_score || 0).toFixed(2)}%</p>
                    </div>
                  </div>
                )}

                {model.file_size && (
                  <div className="flex items-start gap-3">
                    <Package className="w-5 h-5 text-muted-foreground mt-0.5" />
                    <div>
                      <p className="text-sm font-medium">File Size</p>
                      <p className="text-sm text-muted-foreground">{formatFileSize(model.file_size)}</p>
                    </div>
                  </div>
                )}

                <div className="flex items-start gap-3">
                  <Shield className="w-5 h-5 text-muted-foreground mt-0.5" />
                  <div>
                    <p className="text-sm font-medium">License</p>
                    <p className="text-sm text-muted-foreground">{model.license_type || "Not specified"}</p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Training Statistics - Compact */}
          {model.accuracy_score != null && model.accuracy_score !== undefined && (
            <Card className="bg-gradient-card border-border shadow-card">
              <CardHeader className="pb-3">
                <CardTitle className="text-lg flex items-center gap-2">
                  <TrendingUp className="w-4 h-4 text-primary" />
                  Training Statistics
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Accuracy Score */}
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-muted-foreground">Accuracy</span>
                    <span className="text-xl font-bold text-primary">
                      {(model.accuracy_score || 0).toFixed(1)}%
                    </span>
                  </div>
                  <div className="w-full bg-secondary rounded-full h-2 overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-primary to-primary/60 rounded-full transition-all"
                      style={{ width: `${Math.min((model.accuracy_score || 0), 100)}%` }}
                    />
                  </div>
                </div>

                {/* Performance Badge */}
                <div className="flex items-center justify-between pt-2 border-t border-border">
                  <span className="text-sm text-muted-foreground">Performance</span>
                  <Badge
                    className={`${
                      model.accuracy_score >= 90
                        ? "bg-green-500/20 text-green-500 border-green-500/30"
                        : model.accuracy_score >= 75
                        ? "bg-blue-500/20 text-blue-500 border-blue-500/30"
                        : model.accuracy_score >= 60
                        ? "bg-yellow-500/20 text-yellow-500 border-yellow-500/30"
                        : "bg-orange-500/20 text-orange-500 border-orange-500/30"
                    }`}
                  >
                    {model.accuracy_score >= 90
                      ? "Excellent"
                      : model.accuracy_score >= 75
                      ? "Good"
                      : model.accuracy_score >= 60
                      ? "Fair"
                      : "Moderate"}
                  </Badge>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>

      {/* Stripe Checkout Dialog */}
      {model && (
        <StripeProvider>
          <StripeCheckout
            open={checkoutOpen}
            onOpenChange={setCheckoutOpen}
            modelName={model.name}
            price={model.price}
            modelId={model.id}
            onSuccess={handleDownload}
          />
        </StripeProvider>
      )}
    </div>
  );
};

export default ModelDetail;
