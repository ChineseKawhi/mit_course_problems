"""
    A SizeAVL implementation:
    An augmented AVL that keeps track of the node with 
    the number of nodes in the subtree rooted at this node
"""

from SizeBST import SizeBSTNode, SizeBST, size

def height(node):
    if node is None:
        return 0
    else:
        return node.height

def update_node(node):
    node.height = max(height(node.left), height(node.right)) + 1
    node.size = len(node.values) + size(node.left) + size(node.right)

class SizeAVLNode(SizeBSTNode):
    """
    A AVLNode which is augmented to keep track of 
    the number of nodes in the subtree rooted at this node.
    """
    def __init__(self, key, parent, values):
        """
        Creates a node.
        
        Args:
            parent: The node's parent.
            k: key of the node.
        """
        super(SizeAVLNode, self).__init__(key, parent, values)
        self.height = 1

    def check_ri(self):
        """
        Checks the BST representation invariant around this node.
    
        Assert is not true if the RI is violated.

        Running Time:
            O(n)
        """
        size = 1
        assert((height(self.right) - height(self.left) < 2) or\
            (height(self.right) - height(self.left) > -2))
        if self.left is not None:
            assert(self.left.key <= self.key)
            assert(self.left.parent is self)
            size += self.left.size
            self.left.check_ri()
        if self.right is not None:
            assert(self.right.key >= self.key)
            assert(self.right.parent is self)
            size += self.right.size
            self.right.check_ri()
        assert(self.size == size)
        
class SizeAVL(SizeBST):
    """
    An augmented AVL that keeps track of the node with 
    the number of nodes in the subtree rooted at this node
    """
    def __init__(self, node_class = SizeBSTNode):
        super(SizeAVL, self).__init__(node_class)
    
    def insert(self, k, value):
        node = super(SizeAVL, self).insert(k, value)
        self.rebalance(node)
        return node

    def remove(self, k, value):
        node = super(SizeAVL, self).remove(k, value)
        if(node is not None):
            self.rebalance(node.parent)

    def rebalance(self, node):
        while(node is not None):
            update_node(node)
            if(height(node.right) - height(node.left) >= 2):
                if(height(node.right.right) > height(node.right.left)):
                    self.left_rotate(node)
                else:
                    self.right_rotate(node.right)
                    self.left_rotate(node)
            elif(height(node.left) - height(node.right) >= 2):
                if(height(node.left.left) > height(node.left.right)):
                    self.right_rotate(node)
                else:
                    self.left_rotate(node.left)
                    self.right_rotate(node)
            node = node.parent

    def left_rotate(self, x):
        y = x.right
        y.parent = x.parent
        if y.parent is None:
            self.root = y
        else:
            if y.parent.right is x:
                y.parent.right = y
            elif y.parent.left is x:
                y.parent.left = y
        x.right = y.left
        if(x.right is not None):
            x.right.parent = x
        x.parent = y
        y.left = x
        update_node(x)
        update_node(y)

    def right_rotate(self, x):
        y = x.left
        y.parent = x.parent
        if y.parent is None:
            self.root = y
        else:
            if y.parent.left is x:
                y.parent.left = y
            elif y.parent.right is x:
                y.parent.right = y
        x.left = y.right
        if x.left is not None:
            x.left.parent = x
        y.right = x
        x.parent = y
        update_node(x)
        update_node(y)

    def check_ri(self):
        """
        Checks the BST representation invariant.
        
        Assert is not true if the RI is violated.
        """
        if self.root is not None:
            assert(self.root.parent is None)
            self.root.check_ri()
        

if __name__ == "__main__":
    tree = SizeAVL(node_class=SizeAVLNode)
    tree.insert(5)

    tree.insert(1)

    tree.insert(2)

    tree.insert(3)

    tree.insert(4)
    
    print(tree.range(1,5))

    print(tree)
    tree.check_ri()

    tree.remove(3)
    print(tree.range(1,5))
    print(tree)
    tree.check_ri()
