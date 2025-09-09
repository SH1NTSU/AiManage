const std = @import("std");

const Node = struct {
    left: ?*Node,
    right: ?*Node,
    value: u16,

    pub fn new(value: u16) Node {
        return .{ .left = null, .right = null, .value = value };
    }
};

const Tree = struct {
    root: ?*Node,
    allocator: *std.mem.Allocator,
    const Self = @This();

    pub fn new(allocator: *std.mem.Allocator) Self {
        return .{ .root = null, .allocator = allocator };
    }

    pub fn add(tree: *Self, value: u16) !void {
        try Self.addRecursive(&tree.root, tree.allocator, value);
    }

    fn addRecursive(node_ptr_ref: *?*Node, allocator: *std.mem.Allocator, value: u16) !void {
        if (node_ptr_ref.* == null) {
            const new_node_ptr = try allocator.create(Node);
            new_node_ptr.* = Node.new(value);
            node_ptr_ref.* = new_node_ptr;
            return;
        }

        const current_node = node_ptr_ref.*.?;

        if (value < current_node.value) {
            try Self.addRecursive(&current_node.left, allocator, value);
        } else if (value > current_node.value) {
            try Self.addRecursive(&current_node.right, allocator, value);
        } else {
            return;
        }
    }

    pub fn deinit(tree: *Self) void {
        Self.deinitRecursive(tree.root, tree.allocator);
        tree.root = null;
    }

    fn deinitRecursive(node: ?*Node, allocator: *std.mem.Allocator) void {
        if (node) |n| {
            Self.deinitRecursive(n.left, allocator);
            Self.deinitRecursive(n.right, allocator);
            allocator.destroy(n);
        }
    }

    pub fn print(tree: *const Self) void {
        std.debug.print("Tree (In-order traversal):\n", .{});
        Self.printRecursive(tree.root, 0);
    }

    fn printRecursive(node: ?*Node, indent: usize) void {
        if (node) |n| {
            Self.printRecursive(n.right, indent + 4);

            var buffer: [64]u8 = undefined;
            const num_spaces = @min(indent, buffer.len);
            @memset(buffer[0..num_spaces], ' ');
            std.debug.print("{s}{d}\n", .{ buffer[0..num_spaces], n.value });

            Self.printRecursive(n.left, indent + 4);
        }
    }
};

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    var allocator = gpa.allocator();

    var tree = Tree.new(&allocator);
    defer tree.deinit();

    std.debug.print("Adding values: 10, 1, 12, 7, 0\n", .{});
    try tree.add(10);
    try tree.add(1);
    try tree.add(12);
    try tree.add(7);
    try tree.add(0);
    try tree.add(1);

    tree.print();
}
