"""
    A SizeBST implementation:
    An augmented BST that keeps track of the node with 
    the number of nodes in the subtree rooted at this node
"""

from BST import BSTNode, BST

def size(node):
    if node is None:
        return 0
    else:
        return node.size

class SizeBSTNode(BSTNode):
    """
    A BSTNode which is augmented to keep track of 
    the number of nodes in the subtree rooted at this node.
    """
    def __init__(self, key, parent, values):
        """
        Creates a node.
        
        Args:
            parent: The node's parent.
            k: key of the node.
        """
        super(SizeBSTNode, self).__init__(key, parent, values)
        self.size = 1

    def insert(self, node):
        """
        Inserts a node into the subtree rooted at this node.
        
        Args:
            node: The node to be inserted.
        
        Returns:
            succeeds or not

        Running Time:
            O(log(n))
        """
        if(node.key == self.key):
            return False
        elif(node.key < self.key):
            if(self.left is None):
                node.parent = self
                self.left = node
                self.size += 1
                return True
            else:
                if(self.left.insert(node) == True):
                    self.size += 1
                    return True
                else:
                    return False
        else:
            if(self.right is None):
                node.parent = self
                self.right = node
                self.size += 1
                return True
            else:
                
                if(self.right.insert(node) == True):
                    self.size += 1
                    return True
                else:
                    return False

    def remove(self):
        """
        Removes and returns this node from the BST.

        Returns:
            The deleted node with key k.

        Running Time:
            O(log(n))
        """
        if(self.left is None or self.right is None):
            if(self is self.parent.left):
                self.parent.left = self.left or self.right
                if self.parent.left is not None:
                    self.parent.left.parent = self.parent
            else:
                self.parent.right = self.left or self.right
                if(self.parent.right is not None):
                    self.parent.right.parent = self.parent
            # update size
            # parent = self.parent
            # while(parent is not None):
            #     parent.size -= 1
            #     parent = parent.parent
            return self
        else:
            next_larger = self.next_larger()
            self.key, next_larger.key = next_larger.key, self.key
            self.size, next_larger.size = next_larger.size, self.size
            self.values, next_larger.values = next_larger.values, self.values
            # update size
            # parent = next_larger.parent
            # while(parent is not None):
            #     parent.size = size(parent.left) + size(parent.right) + len(parent.values)
            #     parent = parent.parent
            # parent = self.parent
            # while(parent is not None):
            #     parent.size = size(parent.left) + size(parent.right) + len(parent.values)
            #     parent = parent.parent
            return next_larger.remove()

    def rank(self, k, count_k = False):
        """
        Count the number of nodes that key is smaller than k
        from the subtree rooted at this node.
        
        Args:
            k: The key of the node we want to find.
            count_k: include the node that key equals k
        
        Returns:
            The node with key k.

        Running Time:
            O(log(n))
        """
        if k == self.key:
            
            if(self.left is None):
                return len(self.values) if count_k else 0
            else:
                return (len(self.values) if count_k else 0) + self.left.size
        elif k < self.key:
            if(self.left is None):
                return 0
            else:
                return self.left.rank(k, count_k)
        else:
            if self.right is None:  
                return self.size
            else:
                return len(self.values) + size(self.left) + self.right.rank(k, count_k)

    def check_ri(self):
        """
        Checks the BST representation invariant around this node.
    
        Assert is not true if the RI is violated.

        Running Time:
            O(n)
        """
        size = len(self.values)
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

class SizeBST(BST):
    """
    An augmented BST that keeps track of the node with 
    the number of nodes in the subtree rooted at this node
    """
    def __init__(self, node_class = SizeBSTNode):
        super(SizeBST, self).__init__(node_class)
    
    def rank(self, k, count_k):
        """
        Count the number of nodes that key is not greater than k.

        Args:
            k: The key of the node we want to find.
            count_k: include the node that key equals k
        
        Returns:
            The node with key k.

        Running Time:
            O(log(n))
        """
        return self.root.rank(k, count_k)

    def range(self, k1, k2):
        """
        Count the number of nodes that key is between k1 and k2.
        k1, k2 included.

        Running Time:
            O(log(n))
        """
        return 0 if self.root is None else (self.root.rank(k2, True) - self.root.rank(k1, False))

    def insert(self, k, value):
        """
        Inserts a node into the BST.
        
        Args:
            k: The key of the node to be inserted.
        
        Returns:
            The node inserted.
        
        Running Time:
            O(log(n))
        """
        node = self.find(k)
        if(node is None):
            node = self.node_class(k, None, [value])
            if(self.root is None):
                self.root = node
            else:
                self.root.insert(node)
        else:
            node.values.append(value)
            node.size += 1
        return node

    def remove(self, k, value):
        """
        Removes and returns the node with key k from the BST.

        Args:
            k: The key of the node that we want to delete.
            
        Returns:
            The deleted node with key k.

        Running Time:
            O(log(n))
        """
        node = self.find(k)
        if node is None:
            return None
        if(len(node.values)>1):
            node.values.remove(value)
            node.size -= 1
            return node
        else:
            if(node is self.root):
                pseudoroot = self.node_class(0, None, [])
                pseudoroot.left = self.root
                self.root.parent = pseudoroot
                root = self.root.remove()
                self.root = pseudoroot.left
                # wrong first time:
                # the root may be the only one node
                # have to check None
                if self.root is not None:
                    self.root.parent = None
                return root
            else:
                return node.remove()

    def check_ri(self):
        """
        Checks the BST representation invariant.
        
        Assert is not true if the RI is violated.
        """
        if self.root is not None:
            assert(self.root.parent is None)
            self.root.check_ri()

if __name__ == "__main__":
    tree = SizeBST(node_class=SizeBSTNode)
    tree.insert(5, "A")

    tree.insert(1, "A")

    tree.insert(5, "A")

    tree.insert(3, "A")

    tree.insert(4, "A")
    
    print(tree.range(1,5))

    #print(tree)
    tree.check_ri()

    tree.remove(5, "A")
    print(tree.range(1,5))
    print(tree)
    tree.check_ri()
