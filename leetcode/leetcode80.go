// 思路：
// 0 0 1 1 2 1 1 3 3 3 4 5 
// 0 0 1 1 2 3 1 1 3 3 4 5
// 0 0 1 1 2 3 3 1 1 3 4 5
// 0 0 1 1 2 3 3 3 1 1 4 5
// 0 0 1 1 2 3 3 3 4 1 1 5
// 0 0 1 1 2 3 3 3 4 5 1 1
// 从以上交换路径，我们知道这种方式不能一次循环去除所有出现次数超过2次的部分，但是一次循环可以去除一个数的超过2次的部分。
// 因此可以多次循环，每次解决一个数，O(n^2)解决这个问题，好在题目没有要求时间复杂度。

// 因此，方法一：一个数等于2个位置前的数，就是重复超过2次。双层循环，里层循环是：双指针+局部变量：指向当前位置的指针，指向前面待交换位置的指针；
// 一个 记录当前要解决的是哪个数 的局部变量，一个 记录这个数出现了多少次 的变量（用于知道里层循环结束的位置，就也是删除重复后数组的end）。
// 每次循环解决一个数，O(n^2)解决。
func removeDuplicates(nums []int) int {
    // 原数组最大的数，用于外层循环判断是否结束
    oriMaxNum := nums[len(nums)-1]
    // 上次里层循环处理的数，用于外层循环判断是否结束
    lastCircleHandleNum := nums[0]-1
    // 上次里层循环结束后，当前的结束index
    curEndIndex := len(nums)-1
    // 外层步骤：循环，直到上次解决的是原数组中最大的数
    for lastCircleHandleNum < oriMaxNum {
        lastCircleHandleNum, curEndIndex = HandleOneOverDupNum(nums, curEndIndex)
    }
    return curEndIndex+1
}

func HandleOneOverDupNum(nums []int, endIndex int) (int, int) {
    // 长度小于3的不需要处理
    // 指向当前位置的索引
    curIndex := 2
    if curIndex > endIndex {
        return nums[endIndex], endIndex
    }
    // 里层2个步骤：1.先找到第一个重复超过2次的数;2.通过交换删除多余的重复，及找到新的数组结束位置
    findDupFlag := false
    // 后面用于交换的索引
    toSwapIndex := 0
    // 本次处理的重复的数
    thisTurnDupNum := nums[0]
    // 重复的次数
    dupNumExistTimes := 3
    for ; curIndex <= endIndex; curIndex++ {
        if nums[curIndex] == nums[curIndex-2] {
            findDupFlag = true
            toSwapIndex = curIndex
            thisTurnDupNum = nums[curIndex]
            break
        }
    }
    // 没找到重复超过2次的数，直接结束
    if !findDupFlag {
        return nums[endIndex], endIndex
    }

    // 找到重复的数了
    for i := curIndex+1;i<=endIndex;i++ {
        // 如果当前还是重复的这个数
        if nums[i] == thisTurnDupNum {
            dupNumExistTimes++
        } else { // 如果不是这个重复数了，交换
            temp := nums[i]
            nums[i] = nums[toSwapIndex]
            nums[toSwapIndex] = temp

            // 维护待交换的位置
            toSwapIndex++
        }
    }

    return thisTurnDupNum, endIndex-(dupNumExistTimes-2)
}