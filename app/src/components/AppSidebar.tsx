import { Brain, BarChart3, Settings, Store, User, Sparkles } from "lucide-react";
import { NavLink } from "react-router-dom";
import { useEffect, useState } from "react";
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";

const items = [
  { title: "Models", url: "/", icon: Brain },
  { title: "Community", url: "/community", icon: Store },
  { title: "HuggingFace", url: "/huggingface", icon: Sparkles },
  { title: "Statistics", url: "/statistics", icon: BarChart3 },
  { title: "Settings", url: "/settings", icon: Settings },
];

export function AppSidebar() {
  const { open } = useSidebar();
  const [user, setUser] = useState<{ username: string; email: string } | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const token = localStorage.getItem("token");
        if (!token) return;

        const response = await fetch("http://localhost:8081/v1/me", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (response.ok) {
          const data = await response.json();
          setUser(data);
        }
      } catch (error) {
        console.error("Failed to fetch user info:", error);
      }
    };

    fetchUser();
  }, []);

  return (
    <Sidebar className={open ? "w-60" : "w-16"} collapsible="icon">
      <SidebarContent className="flex flex-col h-full">
        {/* Logo - Smaller when collapsed */}
        <div className={`${open ? "p-4" : "p-3"} flex items-center justify-center border-b border-border`}>
          <div className={`${open ? "w-10 h-10" : "w-8 h-8"} rounded-lg bg-gradient-primary flex items-center justify-center shrink-0`}>
            <Brain className={`${open ? "w-6 h-6" : "w-5 h-5"} text-primary-foreground`} />
          </div>
        </div>
        <SidebarGroup className="flex-1">
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <NavLink
                      to={item.url}
                      end
                      className={({ isActive }) =>
                        `flex items-center gap-3 px-3 py-2 rounded-lg transition-all ${
                          isActive
                            ? "bg-accent text-accent-foreground shadow-glow"
                            : "hover:bg-muted text-muted-foreground hover:text-foreground"
                        }`
                      }
                    >
                      <item.icon className="w-5 h-5 shrink-0" />
                      {open && <span className="font-medium">{item.title}</span>}
                    </NavLink>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        {/* User Profile Section - Show only picture when collapsed, smaller when collapsed */}
        {user && (
          <div className={`${open ? "p-4" : "p-3"} border-t border-border flex items-center justify-center`}>
            <div className={`${open ? "w-10 h-10" : "w-8 h-8"} rounded-full bg-gradient-to-br from-primary/20 to-secondary/20 flex items-center justify-center shrink-0`}>
              <User className={`${open ? "w-5 h-5" : "w-4 h-4"} text-primary`} />
            </div>
            {open && (
              <div className="flex flex-col min-w-0 flex-1 ml-3">
                <span className="text-sm font-medium text-foreground truncate">
                  {user.username}
                </span>
                <span className="text-xs text-muted-foreground truncate">
                  {user.email}
                </span>
              </div>
            )}
          </div>
        )}
      </SidebarContent>
    </Sidebar>
  );
}
