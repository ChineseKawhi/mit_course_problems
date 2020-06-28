import rubik

def shortest_path(start, end):
    """
    Using 2-way BFS, finds the shortest path from start_position to
    end_position. Returns a list of moves. 

    You can use the rubik.quarter_twists move set.
    Each move can be applied using rubik.perm_apply
    """
    if(start==end):
        return []
    res_left = {
        start: (None, None)
    }
    res_right = {
        end: (None, None)
    }
    frontier_left = [start]
    frontier_right = [end]
    # Better Answer: count is less than 7 (half the diameter)
    count = 0
    while(len(frontier_left) > 0 and len(frontier_right) > 0 and count < 7):
        next_left = []
        next_right = []
        for cur in frontier_left:
            for rotate in rubik.quarter_twists:
                one_res = rubik.perm_apply(rotate,  cur)
                if(not res_left.has_key(one_res)):
                    next_left.append(one_res)
                    res_left[one_res] = (cur, rotate)
                if(res_right.has_key(one_res)):
                    return compute_res(res_left, res_right, start, end, one_res)
        for cur in frontier_right:
            for rotate in rubik.quarter_twists:
                one_res = rubik.perm_apply(rotate,  cur)
                if(not res_right.has_key(one_res)):
                    next_right.append(one_res)
                    res_right[one_res] = (cur, rubik.perm_inverse(rotate))
                if(res_left.has_key(one_res)):
                    return compute_res(res_left, res_right, start, end, one_res)
        frontier_left = next_left
        frontier_right = next_right
        count +=1
    return None

def compute_res(set_left, set_right, start, end, cur):
    res = []
    cur_left = cur
    while(True):
        if(cur_left == start):
            res.reverse()
            break
        next_value = set_left[cur_left]
        if(next_value[1] != None):
            res.append(next_value[1])
        cur_left = next_value[0] 
    cur_right = cur
    while(True):
        if(cur_right == end):
            break
        next_value = set_right[cur_right]
        if(next_value[1] != None):
            res.append(next_value[1])
        cur_right = next_value[0]
    return res

