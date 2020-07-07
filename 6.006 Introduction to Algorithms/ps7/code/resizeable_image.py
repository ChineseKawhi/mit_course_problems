import imagematrix

class ResizeableImage(imagematrix.ImageMatrix):

    def best_seam(self):
        #print(self.keys())
        parent = {}
        memo = [[0 for __ in range(self.width)] for _ in range(self.height)]
        for i in range(0, self.height):
            for j in range(0, self.width):
                jj = j
                memo[i][j] = memo[i-1][j] if i > 1 else 0
                if(j > 1 and memo[i-1][j-1] < memo[i][j]):
                    memo[i][j] = memo[i-1][j-1]
                    jj = j-1
                if(j < self.width - 1 and memo[i-1][j+1] < memo[i][j]):
                    memo[i][j] = memo[i-1][j+1]
                    jj = j + 1
                memo[i][j] += self.energy(j,i)
                parent[(j,i)] = (jj,i-1)
        start = 0
        min_energy = memo[-1][0]
        for j in range(1, self.width):
            if(memo[-1][j] < min_energy):
                min_energy = memo[-1][j]
                start = j
        cur = (start,self.height-1)
        res = []
        for _ in range(self.height):
            res.append(cur)
            p = parent[cur]
            cur = p
        res.reverse()
        return res

    def remove_best_seam(self):
        self.remove_seam(self.best_seam())
